package xmlstruct_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/twpayne/go-xmlstruct"
)

func TestGenerator(t *testing.T) {
	for _, tc := range []struct {
		name        string
		xmlStrs     []string
		options     []xmlstruct.GeneratorOption
		expectedStr string
	}{
		{
			name: "simple_string",
			xmlStrs: []string{
				joinLines(
					"<a>",
					"  <b>c</b>",
					"</a>",
				),
			},
			expectedStr: joinLines(
				`package main`,
				``,
				`import "encoding/xml"`,
				``,
				`type A struct {`,
				"\tXMLName xml.Name `xml:\"a\"`",
				"\tB       string   `xml:\"b\"`",
				`}`,
			),
		},
		{
			name: "simple_int",
			xmlStrs: []string{
				joinLines(
					"<a>",
					"  <b>2</b>",
					"</a>",
				),
			},
			expectedStr: joinLines(
				`package main`,
				``,
				`import "encoding/xml"`,
				``,
				`type A struct {`,
				"\tXMLName xml.Name `xml:\"a\"`",
				"\tB       int      `xml:\"b\"`",
				`}`,
			),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			generator := xmlstruct.NewGenerator(tc.options...)
			for _, xmlStr := range tc.xmlStrs {
				require.NoError(t, generator.ObserveReader(strings.NewReader(xmlStr)))
			}
			actual, err := generator.Generate()
			require.NoError(t, err)
			assert.Equal(t, tc.expectedStr, string(actual))
		})
	}
}

func joinLines(lines ...string) string {
	return strings.Join(lines, "\n") + "\n"
}