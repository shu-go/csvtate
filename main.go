package main

import (
	"encoding/csv"
	"errors"
	"io"
	"log/slog"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/shu-go/gli/v2"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

type schema struct {
	columns []column
	rep     int
}

type column struct {
	pos  int
	name string

	rep       int
	reppos    []int
	origNames []string
}

func makeSchema(header, excludes []string) *schema {
	s := schema{
		columns: make([]column, 0, len(header)),
	}

	repRE := regexp.MustCompile(`(.+)(\d+)`)
	for i, h := range header {
		h = strings.TrimSpace(h)

		excluded := false
		for _, x := range excludes {
			if strings.Contains(h, x) {
				excluded = true
				break
			}
		}
		if excluded {
			s.columns = append(s.columns, column{
				pos:  i,
				name: h,
			})
			continue
		}

		sub := repRE.FindStringSubmatch(h)

		if len(sub) == 0 {
			s.columns = append(s.columns, column{
				pos:  i,
				name: h,
			})
			continue
		}

		repname := sub[1]
		repnum, err := strconv.Atoi(sub[2])
		if err != nil {
			slog.Error("repeat number is not a number",
				slog.String("error", err.Error()),
				slog.String("column", h),
				slog.String("name", repname),
				slog.String("number", sub[2]))
			return nil
		}

		colidx := s.findByName(repname)
		if colidx == -1 {
			s.columns = append(s.columns, column{
				pos:       i,
				name:      repname,
				rep:       1,
				reppos:    []int{i},
				origNames: []string{h},
			})
		} else {
			col := s.columns[colidx]
			col.rep = repnum
			col.reppos = append(col.reppos, i)
			col.origNames = append(col.origNames, h)
			s.columns[colidx] = col
		}
		if s.rep < repnum {
			s.rep = repnum
		}
	}

	// demote
	demoted := []int{} // index desc
	for i := len(s.columns) - 1; i >= 0; i-- {
		rep := s.columns[i].rep
		if rep > 0 && rep < s.rep {
			demoted = append(demoted, i)
		}
	}
	for _, di := range demoted {
		d := s.columns[di]
		for j := len(d.reppos) - 1; j > 0; j-- {
			s.columns = append(s.columns, column{
				pos:  d.reppos[j],
				name: d.origNames[j],
				rep:  0,
			})
		}
		d.pos = d.reppos[0]
		d.name = d.origNames[0]
		d.rep = 0
		d.reppos = nil
		d.origNames = nil
		s.columns[di] = d
	}

	return &s
}

func (s schema) findByName(name string) int {
	for i, c := range s.columns {
		if c.name == name {
			return i
		}
	}
	return -1
}

func (s schema) repEmpty(record []string, repidx int, repeatIf string) bool {
	if repeatIf == "any" {
		// false (NOT EMPTY) if AT LEAST ONE column is NOT empty.
		// true (EMPTY) if ALL columns are empty
		for _, sc := range s.columns {
			if sc.rep <= repidx {
				continue
			}

			if record[sc.reppos[repidx]] != "" {
				return false
			}
		}
		return true
	}

	// all
	// false (NOT EMPTY) if ALL column are NOT empty.
	// true (EMPTY) if AT LEAST ONE column is empty
	for _, sc := range s.columns {
		if sc.rep <= repidx {
			continue
		}

		if record[sc.reppos[repidx]] == "" {
			return true
		}
	}
	return false
}

type globalCmd struct {
	Encoding string `type:"choice" choices:"sjis,utf8" default:"sjis" help:"sjis or utf8"`
	sjis     bool

	Header bool `cli:"header" default:"true" help:"output header record"`

	Excludes []string `cli:"exclude" help:"do not normalize about the columns"`
	RepeatIf string   `cli:"repeat-if" type:"choice" choices:"all,any" default:"any" help:"normalize if [all | any] columns are non-empty"`
}

func (c *globalCmd) Before() {
	c.sjis = strings.Contains(strings.ToLower(c.Encoding), "jis")
	slog.Info("sjis", slog.String("--encoding", c.Encoding))
}

func (c globalCmd) Run(args []string) error {
	if len(args) < 1 {
		return errors.New("no input")
	}
	infile, err := os.Open(args[0])
	if err != nil {
		return err
	}
	defer infile.Close()

	var filereader io.Reader = infile
	if c.sjis {
		filereader = transform.NewReader(infile, japanese.ShiftJIS.NewDecoder())
	}

	r := csv.NewReader(filereader)
	r.ReuseRecord = true
	r.TrimLeadingSpace = true

	records, err := Convert(r, c.Header, c.Excludes, c.RepeatIf)
	if err != nil {
		return err
	}

	var filewriter io.Writer
	if len(args) >= 2 {
		outfile, err := os.Create(args[1])
		if err != nil {
			return err
		}
		defer outfile.Close()

		filewriter = outfile
	} else {
		filewriter = os.Stdout
	}

	if c.sjis {
		filewriter = transform.NewWriter(filewriter, japanese.ShiftJIS.NewEncoder())
	}

	w := csv.NewWriter(filewriter)
	err = w.WriteAll(records)
	if err != nil {
		return err
	}

	return nil
}

func Convert(r *csv.Reader, outputHeader bool, excludes []string, repeatIf string) ([][]string, error) {
	header, err := r.Read()
	if len(header) == 0 {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	slog.Debug("read header", slog.Any("header", header))

	s := makeSchema(header, excludes)
	slog.Debug("makeSchema", slog.Any("schema", s))

	records := make([][]string, 0, 100)

	if outputHeader {
		// header
		hrecord := make([]string, len(s.columns))
		for i, sc := range s.columns {
			hrecord[i] = sc.name
		}
		records = append(records, hrecord)
	}

	// body
	for {
		slog.Debug("loop")

		inrecord, err := r.Read()
		slog.Debug("read", slog.Any("record", inrecord), slog.Any("error", err))
		if len(inrecord) == 0 {
			break
		}
		if err != nil {
			return nil, err
		}

		if s.rep < 2 {
			outrecord := make([]string, len(s.columns))
			for i, sc := range s.columns {
				slog.Debug("rep0", slog.String("name", sc.name))
				outrecord[i] = inrecord[sc.pos]
			}
			records = append(records, outrecord)
			continue
		}

		for repidx := 0; repidx < s.rep; repidx++ {
			slog.Debug("", slog.Int("repidx", repidx))
			if s.repEmpty(inrecord, repidx, repeatIf) {
				slog.Debug("empty", slog.Int("rep", s.rep))
				break
			}

			outrecord := make([]string, len(s.columns))
			for i, sc := range s.columns {
				if sc.rep > 0 {
					slog.Debug("rep", slog.String("name", sc.name), slog.Int("rep", sc.rep), slog.Any("reppos", sc.reppos))
					outrecord[i] = inrecord[sc.reppos[repidx]]
				} else {
					slog.Debug("rep0", slog.String("name", sc.name))
					outrecord[i] = inrecord[sc.pos]
				}
			}
			records = append(records, outrecord)
		}
	}

	return records, nil
}

// Version is app version
var Version string

func main() {
	app := gli.NewWith(&globalCmd{})
	app.Name = "csvtate"
	app.Desc = "normalize repetitive columns"
	app.Version = Version
	app.Usage = `csvtate repetitive_input.csv normalized_output.csv

a,b1,b2,c1,c2
1,2,3,4,5
 |
 v
a,b,c
1,2,4
1,3,5`
	app.Copyright = "(C) 2024 Shuhei Kubota"
	if err := app.Run(os.Args); err != nil {
		os.Exit(1)
	}

}
