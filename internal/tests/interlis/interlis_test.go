package interlis

import (
	"encoding/xml"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/twpayne/go-xmlstruct"
)

func TestInterlis(t *testing.T) {
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

	require.NoError(t, generator.ObserveFile("testdata/metadata_gm03.xml"))

	actualSource, err := generator.Generate()
	require.NoError(t, err)

	require.NoError(t, os.WriteFile("interlis.gen.go.actual", actualSource, 0o666))

	expectedSource, err := os.ReadFile("interlis.gen.go")
	require.NoError(t, err)
	require.Equal(t, string(expectedSource), string(actualSource))

	data, err := os.ReadFile("testdata/metadata_gm03.xml")
	require.NoError(t, err)
	var transfer Transfer
	require.NoError(t, xml.Unmarshal(data, &transfer))

	assert.Equal(t, "geocat.ch", transfer.HeaderSection.Sender)
}
