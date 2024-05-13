package main_test

import (
	"bytes"
	"encoding/csv"
	"log"
	"log/slog"
	"os"
	"testing"

	"github.com/shu-go/gotwant"
	"github.com/shu-go/shandler/color"
	"github.com/shu-go/shandler/opt"

	main "github.com/shu-go/csvtate"
)

var h *opt.OptHandler

func init() {
	h = opt.NewHandler(func(opts *slog.HandlerOptions) slog.Handler {
		copts := &color.HandlerOptions{}
		if opts != nil {
			copts.Level = opts.Level
		}
		return color.NewHandler(os.Stderr, copts, nil)
	}, nil)
	slog.SetDefault(slog.New(h))
	log.SetFlags(0)
}

func TestConvert(t *testing.T) {
	t.Run("NoTate", func(t *testing.T) {
		in := bytes.NewBufferString(`a,b,c
1,2,3`)
		records, err := main.Convert(csv.NewReader(in), true, nil, "any")

		gotwant.TestError(t, err, nil)
		gotwant.Test(t, records, [][]string{
			{"a", "b", "c"},
			{"1", "2", "3"},
		})

		in = bytes.NewBufferString(`a,b1,b2,c
1,2,3,4`)
		records, err = main.Convert(csv.NewReader(in), true, []string{"b"}, "any")

		gotwant.TestError(t, err, nil)
		gotwant.Test(t, records, [][]string{
			{"a", "b1", "b2", "c"},
			{"1", "2", "3", "4"},
		})
	})

	t.Run("One", func(t *testing.T) {
		in := bytes.NewBufferString(`a,b1,c
1,2,3`)
		records, err := main.Convert(csv.NewReader(in), true, nil, "any")

		gotwant.TestError(t, err, nil)
		gotwant.Test(t, records, [][]string{
			{"a", "b", "c"},
			{"1", "2", "3"},
		})

		in = bytes.NewBufferString(`a,b1,c1
1,2,3`)
		records, err = main.Convert(csv.NewReader(in), true, nil, "any")

		gotwant.TestError(t, err, nil)
		gotwant.Test(t, records, [][]string{
			{"a", "b", "c"},
			{"1", "2", "3"},
		})
	})

	t.Run("Two", func(t *testing.T) {

		in := bytes.NewBufferString(`a,b1,b2,c
1,2,3,4`)
		records, err := main.Convert(csv.NewReader(in), true, nil, "any")

		gotwant.TestError(t, err, nil)
		gotwant.Test(t, records, [][]string{
			{"a", "b", "c"},
			{"1", "2", "4"},
			{"1", "3", "4"},
		})

		in = bytes.NewBufferString(`a,b1,c1,b2,c2
1,2,3,4,5`)
		records, err = main.Convert(csv.NewReader(in), true, nil, "any")

		gotwant.TestError(t, err, nil)
		gotwant.Test(t, records, [][]string{
			{"a", "b", "c"},
			{"1", "2", "3"},
			{"1", "4", "5"},
		})
	})

	t.Run("Three", func(t *testing.T) {
		in := bytes.NewBufferString(`a1,b1,c1
1,2,3`)
		records, err := main.Convert(csv.NewReader(in), true, nil, "any")

		gotwant.TestError(t, err, nil)
		gotwant.Test(t, records, [][]string{
			{"a", "b", "c"},
			{"1", "2", "3"},
		})

		in = bytes.NewBufferString(`a1,b1,c1,a2,b2,c2
1,2,3,4,5,6`)
		records, err = main.Convert(csv.NewReader(in), true, nil, "any")

		gotwant.TestError(t, err, nil)
		gotwant.Test(t, records, [][]string{
			{"a", "b", "c"},
			{"1", "2", "3"},
			{"4", "5", "6"},
		})
	})
}

func TestDemoted(t *testing.T) {
	t.Run("LessRep", func(t *testing.T) {
		in := bytes.NewBufferString(`a,b1,b2,b3,c1,c2,c3,d1
1,2,3,4,5,6,7,8`)
		records, err := main.Convert(csv.NewReader(in), true, nil, "any")

		gotwant.TestError(t, err, nil)
		gotwant.Test(t, records, [][]string{
			{"a", "b", "c", "d1"},
			{"1", "2", "5", "8"},
			{"1", "3", "6", "8"},
			{"1", "4", "7", "8"},
		})
	})

	t.Run("Exclude", func(t *testing.T) {
		in := bytes.NewBufferString(`a,b1,b2,b3,c1,c2,c3,d1
1,2,3,4,5,6,7,8`)
		records, err := main.Convert(csv.NewReader(in), true, []string{"c"}, "any")

		gotwant.TestError(t, err, nil)
		gotwant.Test(t, records, [][]string{
			{"a", "b", "c1", "c2", "c3", "d1"},
			{"1", "2", "5", "6", "7", "8"},
			{"1", "3", "5", "6", "7", "8"},
			{"1", "4", "5", "6", "7", "8"},
		})
	})
}

func TestRepeatIf(t *testing.T) {
	t.Run("any", func(t *testing.T) {
		in := bytes.NewBufferString(`a,b1,b2,b3,b4,c1,c2,c3,c4,d1
1,2,3,,,6,,8,,10`)
		records, err := main.Convert(csv.NewReader(in), true, nil, "any")

		gotwant.TestError(t, err, nil)
		gotwant.Test(t, records, [][]string{
			{"a", "b", "c", "d1"},
			{"1", "2", "6", "10"},
			{"1", "3", "", "10"},
			{"1", "", "8", "10"},
		})
	})

	t.Run("all", func(t *testing.T) {
		in := bytes.NewBufferString(`a,b1,b2,b3,b4,c1,c2,c3,c4,d1
1,2,3,,,6,,8,,10`)
		records, err := main.Convert(csv.NewReader(in), true, nil, "all")

		gotwant.TestError(t, err, nil)
		gotwant.Test(t, records, [][]string{
			{"a", "b", "c", "d1"},
			{"1", "2", "6", "10"},
		})
	})
}
