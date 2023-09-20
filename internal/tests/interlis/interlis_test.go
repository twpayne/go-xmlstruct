package interlis

import (
	"encoding/xml"
	"os"
	"testing"

	"github.com/alecthomas/assert/v2"

	"github.com/twpayne/go-xmlstruct"
)

func TestInterlis(t *testing.T) {
	t.Parallel()

	generator := xmlstruct.NewGenerator(
		xmlstruct.WithExportRenames(map[string]string{
			"BOUNDARY":      "Boundary",
			"COORD":         "Coord",
			"DATASECTION":   "DataSection",
			"HEADERSECTION": "HeaderSection",
			"MODELS":        "Models",
			"POLYLINE":      "PolyLine",
			"SENDER":        "Sender",
			"SURFACE":       "Surface",
			"TRANSFER":      "Transfer",
			"VERSION":       "Version",
		}),
		xmlstruct.WithNamedTypes(true),
		xmlstruct.WithPackageName("interlis"),
		xmlstruct.WithPreserveOrder(true),
	)

	assert.NoError(t, generator.ObserveFile("testdata/metadata_gm03.xml"))

	actualSource, err := generator.Generate()
	assert.NoError(t, err)

	assert.NoError(t, os.WriteFile("interlis.gen.go.actual", actualSource, 0o666))

	expectedSource, err := os.ReadFile("interlis.gen.go")
	assert.NoError(t, err)
	assert.Equal(t, string(expectedSource), string(actualSource))

	data, err := os.ReadFile("testdata/metadata_gm03.xml")
	assert.NoError(t, err)
	var transfer Transfer
	assert.NoError(t, xml.Unmarshal(data, &transfer))

	assert.Equal(t, "geocat.ch", transfer.HeaderSection.Sender)
}
