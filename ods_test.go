package openoffice

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestODSFile_ParseContent(t *testing.T) {
	f, err := OpenODS("./testdata/test.ods")
	assert.Nil(t, err)
	defer f.Close()

	doc, err := f.ParseContent()
	assert.Nil(t, err)

	expected := [][]string{
		{"A", "1", "A cell containing\nmore than one line."},
		{"B", "foo"},
		{"", "4", "quote\"quote"},
		{"14.01.12"},
		{},
		{"cell spanning two columns", ""},
		{},
		{},
		{"aaa", "cell spanning two rows", "ccc"},
		{"aa", "", "cc"},
		{},
		{"same content", "same content", "same content"},
		{},
		{"same content"},
		{"same content"},
		{"same content"},
		{},
		{"Cell with inline styles"},
	}

	for i, row := range doc.Table[0].Strings() {
		assert.Equal(t, row, expected[i])
	}
}
