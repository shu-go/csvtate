normalize repetitive columns

[![Go Report Card](https://goreportcard.com/badge/github.com/shu-go/csvtate)](https://goreportcard.com/report/github.com/shu-go/csvtate)
![MIT License](https://img.shields.io/badge/License-MIT-blue)

# csvtate

```sh
csvtate repetitive_input.csv normalized_output.csv
```

```csv
a,b1,b2,c1,c2
1,2,3,4,5
 |
 v
a,b,c
1,2,4
1,3,5
```

# Install

## GitHub Releases

https://github.com/shu-go/csvtate/releases

## Go install

```sh
go install github.com/shu-go/csvtate@latest
```
