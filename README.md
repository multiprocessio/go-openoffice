# go-openoffice

A Go library for reading OpenOffice/LibreOffice .ods (and .odf) files.

## Example

```go
$ cat ./examples/dump/main.go
package main

import "fmt"
import "strings"
import "github.com/multiprocessio/go-openoffice"

func main() {
	f, err := openoffice.OpenODS("testdata/test.ods")
	if err != nil {
		panic(err)
	}

	doc, err := f.ParseContent()
	if err != nil {
		panic(err)
	}

	for _, t := range doc.Sheets {
		fmt.Printf("Sheet Name: %s, Rows: %d\n", t.Name, len(t.Rows))

		for _, row := range t.Strings() {
			fmt.Println(strings.Join(row, ","))
		}

		fmt.Println()
	}
}
```

## History

This project is a fork of https://github.com/knieriem/odf.