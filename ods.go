// This package implements rudimentary support
// for reading Open Document Spreadsheet files. At current
// stage table data can be accessed.
package openoffice

import (
	"bytes"
	"encoding/xml"
	"errors"
	"io"
	"strconv"
	"strings"
)

type Doc struct {
	XMLName xml.Name `xml:"document-content"`
	Sheets   []Sheet  `xml:"body>spreadsheet>table"`
}

type Sheet struct {
	Name   string   `xml:"name,attr"`
	Column []string `xml:"table-column"`
	Rows    []Row    `xml:"table-row"`
}

type Row struct {
	RepeatedRows int `xml:"number-rows-repeated,attr"`

	Cells []Cell `xml:",any"` // use ",any" to match table-cell and covered-table-cell
}

func (r *Row) IsEmpty() bool {
	for _, c := range r.Cells {
		if !c.IsEmpty() {
			return false
		}
	}
	return true
}

// Return the contents of a row as a slice of strings. Cells that are
// covered by other cells will appear as empty strings.
func (r *Row) Strings(b *bytes.Buffer) (row []string) {
	n := len(r.Cells)
	if n == 0 {
		return
	}

	// remove trailing empty cells
	for i := n - 1; i >= 0; i-- {
		if !r.Cells[i].IsEmpty() {
			break
		}
		n--
	}
	r.Cells = r.Cells[:n]

	n = 0
	// calculate the real number of cells (including repeated)
	for _, c := range r.Cells {
		switch {
		case c.RepeatedCols != 0:
			n += c.RepeatedCols
		default:
			n++
		}
	}

	row = make([]string, n)
	w := 0
	for _, c := range r.Cells {
		cs := ""
		if c.XMLName.Local != "covered-table-cell" {
			cs = c.PlainText(b)
		}
		row[w] = cs
		w++
		switch {
		case c.RepeatedCols != 0:
			for j := 1; j < c.RepeatedCols; j++ {
				row[w] = cs
				w++
			}
		}
	}
	return
}

type Cell struct {
	XMLName xml.Name

	// attributes
	ValueType    string `xml:"value-type,attr"`
	Value        string `xml:"value,attr"`
	Formula      string `xml:"formula,attr"`
	RepeatedCols int    `xml:"number-columns-repeated,attr"`
	ColSpan      int    `xml:"number-columns-spanned,attr"`

	P []Par `xml:"p"`
}

func (c *Cell) IsEmpty() (empty bool) {
	switch len(c.P) {
	case 0:
		empty = true
	case 1:
		if c.P[0].XML == "" {
			empty = true
		}
	}
	return
}

// PlainText extracts the text from a cell. Space tags (<text:s text:c="#">)
// are recognized. Inline elements (like span) are ignored, but the
// text they contain is preserved
func (c *Cell) PlainText(b *bytes.Buffer) string {
	n := len(c.P)
	if n == 1 {
		return c.P[0].PlainText(b)
	}

	b.Reset()
	for i := range c.P {
		if i != n-1 {
			c.P[i].writePlainText(b)
			b.WriteByte('\n')
		} else {
			c.P[i].writePlainText(b)
		}
	}
	return b.String()
}

type Par struct {
	XML string `xml:",innerxml"`
}

func (p *Par) PlainText(b *bytes.Buffer) string {
	for i := range p.XML {
		if p.XML[i] == '<' || p.XML[i] == '&' {
			b.Reset()
			p.writePlainText(b)
			return b.String()
		}
	}
	return p.XML
}
func (p *Par) writePlainText(b *bytes.Buffer) {
	for i := range p.XML {
		if p.XML[i] == '<' || p.XML[i] == '&' {
			goto decode
		}
	}
	b.WriteString(p.XML)
	return

decode:
	d := xml.NewDecoder(strings.NewReader(p.XML))
	for {
		t, _ := d.Token()
		if t == nil {
			break
		}
		switch el := t.(type) {
		case xml.StartElement:
			switch el.Name.Local {
			case "s":
				n := 1
				for _, a := range el.Attr {
					if a.Name.Local == "c" {
						n, _ = strconv.Atoi(a.Value)
					}
				}
				for i := 0; i < n; i++ {
					b.WriteByte(' ')
				}
			}
		case xml.CharData:
			b.Write(el)
		}
	}
}

func (t *Sheet) Width() int {
	return len(t.Column)
}
func (t *Sheet) Height() int {
	return len(t.Rows)
}
func (t *Sheet) Strings() (s [][]string) {
	var b bytes.Buffer

	n := len(t.Rows)
	if n == 0 {
		return
	}

	// remove trailing empty rows
	for i := n - 1; i >= 0; i-- {
		if !t.Rows[i].IsEmpty() {
			break
		}
		n--
	}
	t.Rows = t.Rows[:n]

	n = 0
	// calculate the real number of rows (including repeated rows)
	for _, r := range t.Rows {
		switch {
		case r.RepeatedRows != 0:
			n += r.RepeatedRows
		default:
			n++
		}
	}

	s = make([][]string, n)
	w := 0
	for _, r := range t.Rows {
		row := r.Strings(&b)
		s[w] = row
		w++
		for j := 1; j < r.RepeatedRows; j++ {
			s[w] = row
			w++
		}
	}
	return
}

type ODSFile struct {
	*ODFFile
}

// Open an ODS file. If the file doesn't exist or doesn't look
// like a spreadsheet file, an error is returned.
func OpenODS(fileName string) (*ODSFile, error) {
	f, err := OpenODF(fileName)
	if err != nil {
		return nil, err
	}
	return newODSFile(f)
}

// NewReader initializes a File struct with an already opened
// ODS file, and checks the spreadsheet's media type.
func NewODSReader(r io.ReaderAt, size int64) (*ODSFile, error) {
	f, err := NewODFReader(r, size)
	if err != nil {
		return nil, err
	}
	return newODSFile(f)
}

func newODSFile(f *ODFFile) (*ODSFile, error) {
	if f.MimeType != MimeTypePfx+"spreadsheet" {
		f.Close()
		return nil, errors.New("not a spreadsheet")
	}
	return &ODSFile{f}, nil
}

// Parse the content.xml part of an ODS file. On Success
// the returned Doc will contain the data of the rows and cells
// of the table(s) contained in the ODS file.
func (f *ODSFile) ParseContent() (*Doc, error) {
	content, err := f.Open("content.xml")
	if err != nil {
		return nil, err
	}
	defer content.Close()

	d := xml.NewDecoder(content)
	var doc Doc
	err = d.Decode(&doc)
	return &doc, err
}
