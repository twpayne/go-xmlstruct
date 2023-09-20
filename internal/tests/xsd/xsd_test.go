package xsd

import (
	"bytes"
	"encoding/xml"
	"os"
	"testing"

	"github.com/alecthomas/assert/v2"
	"golang.org/x/net/html/charset"

	"github.com/twpayne/go-xmlstruct"
)

func TestXSD(t *testing.T) {
	t.Parallel()

	generator := xmlstruct.NewGenerator(
		xmlstruct.WithExportRenames(map[string]string{
			"appinfo": "AppInfo",
		}),
		xmlstruct.WithNamedTypes(true),
		xmlstruct.WithPackageName("xsd"),
	)

	filenames := []string{
		"testdata/kml22gx.xsd",
		"testdata/ogckml22.xsd",
		"testdata/xacml-core-v3-schema-wd-17.xsd",
	}

	for _, filename := range filenames {
		assert.NoError(t, generator.ObserveFile(filename))
	}

	actualSource, err := generator.Generate()
	assert.NoError(t, err)

	assert.NoError(t, os.WriteFile("xsd.gen.go.actual", actualSource, 0o666))

	expectedSource, err := os.ReadFile("xsd.gen.go")
	assert.NoError(t, err)
	assert.Equal(t, string(expectedSource), string(actualSource))

	for _, filename := range filenames {
		data, err := os.ReadFile(filename)
		assert.NoError(t, err)

		decoder := xml.NewDecoder(bytes.NewReader(data))
		decoder.CharsetReader = charset.NewReaderLabel

		var schema Schema
		assert.NoError(t, decoder.Decode(&schema))

		switch filename {
		case "testdata/kml22gx.xsd":
			assert.Equal(t, []Import{
				{
					Namespace:      "http://www.opengis.net/kml/2.2",
					SchemaLocation: "http://schemas.opengis.net/kml/2.2.0/ogckml22.xsd",
				},
			}, schema.Import)
		case "testdata/ogckml22.xsd":
			assert.Equal(t, "ogckml22.xsd 2008-01-23", *schema.Annotation.AppInfo)
		case "testdata/xacml-core-v3-schema-wd-17.xsd":
			assert.Equal(t, []Import{
				{
					Namespace:      "http://www.w3.org/XML/1998/namespace",
					SchemaLocation: "http://www.w3.org/2001/xml.xsd",
				},
			}, schema.Import)
		}
	}
}
