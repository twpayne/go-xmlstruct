package gpx

import (
	"bytes"
	"encoding/xml"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/html/charset"

	"github.com/twpayne/go-xmlstruct"
)

func TestGPX(t *testing.T) {
	generator := xmlstruct.NewGenerator(
		xmlstruct.WithExportRenames(map[string]string{
			"gpx":     "GPX",
			"maxlat":  "MaxLat",
			"maxlon":  "MaxLon",
			"minlat":  "MinLat",
			"minlon":  "MinLon",
			"rtept":   "RtePt",
			"trkpt":   "TrkPt",
			"trkseg":  "TrkSeg",
			"url":     "URL",
			"urlname": "URLName",
		}),
		xmlstruct.WithPackageName("gpx"),
	)

	filenames := []string{
		"testdata/ashland.gpx",
		"testdata/fells_loop.gpx",
		"testdata/mystic_basin_trail.gpx",
	}

	for _, filename := range filenames {
		require.NoError(t, generator.ObserveFile(filename))
	}

	actualSource, err := generator.Generate()
	require.NoError(t, err)

	require.NoError(t, os.WriteFile("gpx.gen.go.actual", actualSource, 0o666))

	expectedSource, err := os.ReadFile("gpx.gen.go")
	require.NoError(t, err)
	require.Equal(t, string(expectedSource), string(actualSource))

	for _, filename := range filenames {
		data, err := os.ReadFile(filename)
		require.NoError(t, err)

		decoder := xml.NewDecoder(bytes.NewReader(data))
		decoder.CharsetReader = charset.NewReaderLabel

		var gpx GPX
		require.NoError(t, decoder.Decode(&gpx))

		switch filename {
		case "testdata/ashland.gpx":
			assert.Equal(t, "Vil and Dan", *gpx.Author)
		case "testdata/fells_loop.gpx":
			assert.Equal(t, time.Date(2002, 2, 27, 17, 18, 33, 0, time.UTC), *gpx.Time)
		case "testdata/mystic_basin_trail.gpx":
			assert.Equal(t, "Mystic River Basin Trails", gpx.Metadata.Name)
		}
	}
}
