package musicxml_test

import (
	"archive/zip"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/alecthomas/assert/v2"

	"github.com/twpayne/go-xmlstruct"
)

func TestMusicXML(t *testing.T) {
	t.Parallel()

	generator := xmlstruct.NewGenerator(
		xmlstruct.WithExportRenames(map[string]string{
			"ff":     "FF",
			"halign": "HAlign",
			"mf":     "MF",
			"mp":     "MP",
			"pp":     "PP",
			"sfp":    "SFP",
		}),
		xmlstruct.WithNamedTypes(true),
		xmlstruct.WithPackageName("musicxml"),
	)

	// testdata/xmlsamples.zip contains the MusicXML and corresponding files
	// from https://www.musicxml.com/music-in-musicxml/example-set/.
	file, err := os.Open("testdata/xmlsamples.zip")
	assert.NoError(t, err)
	defer file.Close()

	fileInfo, err := file.Stat()
	assert.NoError(t, err)

	zipReader, err := zip.NewReader(file, fileInfo.Size())
	assert.NoError(t, err)

	utf16TestFiles := map[string]bool{
		"MozaChloSample.musicxml": true,
		"MozaVeilSample.musicxml": true,
	}

	assert.NoError(t, generator.ObserveFS(zipReader, ".", func(path string, _ fs.DirEntry, err error) error {
		switch {
		case err != nil:
			return err
		case path == "__MACOSX":
			return fs.SkipDir
		case filepath.Ext(path) != ".musicxml":
			return xmlstruct.SkipFile
		case utf16TestFiles[path]:
			return xmlstruct.SkipFile
		default:
			return nil
		}
	}))

	actualSource, err := generator.Generate()
	assert.NoError(t, err)

	assert.NoError(t, os.WriteFile("musicxml.gen.go.actual", actualSource, 0o666))

	expectedSource, err := os.ReadFile("musicxml.gen.go")
	assert.NoError(t, err)
	assert.Equal(t, string(expectedSource), string(actualSource))
}
