package osm

import (
	"compress/bzip2"
	"os"
	"testing"

	"github.com/alecthomas/assert/v2"

	"github.com/twpayne/go-xmlstruct"
)

func TestOSM(t *testing.T) {
	generator := xmlstruct.NewGenerator(
		xmlstruct.WithExportRenames(map[string]string{
			"osm":    "OSM",
			"minlat": "MinLat",
			"maxlat": "MaxLat",
			"minlon": "MinLon",
			"maxlon": "MaxLon",
		}),
		xmlstruct.WithNamedTypes(true),
		xmlstruct.WithPackageName("osm"),
		xmlstruct.WithPreserveOrder(true),
	)

	file, err := os.Open("testdata/liechtenstein-latest.osm.bz2")
	assert.NoError(t, err)
	defer file.Close()

	assert.NoError(t, generator.ObserveReader(bzip2.NewReader(file)))

	actualSource, err := generator.Generate()
	assert.NoError(t, err)

	assert.NoError(t, os.WriteFile("osm.gen.go.actual", actualSource, 0o666))

	expectedSource, err := os.ReadFile("osm.gen.go")
	assert.NoError(t, err)
	assert.Equal(t, string(expectedSource), string(actualSource))
}
