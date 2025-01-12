package rss_test

import (
	"encoding/xml"
	"os"
	"testing"

	"github.com/alecthomas/assert/v2"

	"github.com/twpayne/go-xmlstruct"
	"github.com/twpayne/go-xmlstruct/internal/tests/rss"
)

func TestPlay(t *testing.T) {
	t.Parallel()

	generator := xmlstruct.NewGenerator(
		xmlstruct.WithExportRenames(map[string]string{
			"guid": "GUID",
			"rss":  "RSS",
			"url":  "URL",
		}),
		xmlstruct.WithNamedTypes(true),
		xmlstruct.WithPackageName("rss"),
		xmlstruct.WithPreserveOrder(true),
	)

	assert.NoError(t, generator.ObserveFile("testdata/sample-rss-2.xml"))

	actualSource, err := generator.Generate()
	assert.NoError(t, err)

	assert.NoError(t, os.WriteFile("rss.gen.go.actual", actualSource, 0o666))

	expectedSource, err := os.ReadFile("rss.gen.go")
	assert.NoError(t, err)
	assert.Equal(t, string(expectedSource), string(actualSource))

	data, err := os.ReadFile("testdata/sample-rss-2.xml")
	assert.NoError(t, err)

	var nasaSpaceStationNews rss.RSS
	assert.NoError(t, xml.Unmarshal(data, &nasaSpaceStationNews))

	assert.Equal(t, 5, len(nasaSpaceStationNews.Channel.Item))
	assert.Equal(t, "Louisiana Students to Hear from NASA Astronauts Aboard Space Station", *nasaSpaceStationNews.Channel.Item[0].Title)
}
