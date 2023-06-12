package gml

import (
	"archive/zip"
	"encoding/xml"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/alecthomas/assert/v2"

	"github.com/twpayne/go-xmlstruct"
)

func TestGML(t *testing.T) {
	generator := xmlstruct.NewGenerator(
		xmlstruct.WithExportRenames(map[string]string{
			"note": "LowerNote",
		}),
		xmlstruct.WithNameFunc(func(name xml.Name) xml.Name {
			if name.Space != "http://www.opengis.net/gml/3.2" {
				return xml.Name{}
			}
			return name
		}),
		xmlstruct.WithNamedTypes(true),
		xmlstruct.WithPackageName("gml"),
	)

	// testdata/ets-gml32.zip contains the GML 3.2 (ISO 19136:2007) Conformance
	// Test Suite from https://github.com/opengeospatial/ets-gml32.
	file, err := os.Open("testdata/ets-gml32.zip")
	assert.NoError(t, err)
	defer file.Close()

	fileInfo, err := file.Stat()
	assert.NoError(t, err)

	zipReader, err := zip.NewReader(file, fileInfo.Size())
	assert.NoError(t, err)

	assert.NoError(t, generator.ObserveFS(zipReader, "ets-gml32-master/src/test/resources", func(path string, dirEntry fs.DirEntry, err error) error {
		switch {
		case err != nil:
			return err
		case filepath.Ext(path) != ".xml":
			return xmlstruct.SkipFile
		default:
			return nil
		}
	}))

	actualSource, err := generator.Generate()
	assert.NoError(t, err)

	assert.NoError(t, os.WriteFile("gml.gen.go.actual", actualSource, 0o666))

	expectedSource, err := os.ReadFile("gml.gen.go")
	assert.NoError(t, err)
	assert.Equal(t, string(expectedSource), string(actualSource))
}
