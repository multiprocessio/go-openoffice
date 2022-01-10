# go-openoffice

A Go library for reading OpenOffice/LibreOffice .ods (and .odf) files.

## Example

```go
package main

import "fmt"
import "github.com/multiprocessio/go-openoffice"

func main() {
  f, err := openoffice.NewODSReader("spreadsheet.ods")
  if err != nil {
    panic(err)
  }


}
```

## History

This project is a fork of https://github.com/knieriem/odf.