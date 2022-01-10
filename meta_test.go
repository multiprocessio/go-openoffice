package openoffice

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestODFFile_Meta(t *testing.T) {
	f, err := OpenODF("./testdata/test.ods")
	assert.Nil(t, err)
	defer f.Close()

	m, err := f.Meta()
	assert.Nil(t, err)

	time, err := m.Meta.CreationDate.Time()
	assert.Nil(t, err)

	assert.Equal(t, time.Format("2006-01-02"), "2012-01-10")
	assert.Equal(t, m.Meta.Title, "Test Spreadsheet for odf/ods package")
}
