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
				xmlstruct.DefaultHeader,
				``,
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
				xmlstruct.DefaultHeader,
				``,
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
		{
			name: "int_type",
			options: []xmlstruct.GeneratorOption{
				xmlstruct.WithHeader("// Custom header."),
			},
			xmlStrs: []string{
				joinLines(
					"<a>",
					"  <b>c</b>",
					"</a>",
				),
			},
			expectedStr: joinLines(
				`// Custom header.`,
				``,
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
			name: "int_type",
			options: []xmlstruct.GeneratorOption{
				xmlstruct.WithIntType("int64"),
			},
			xmlStrs: []string{
				joinLines(
					"<a>",
					"  <b>2</b>",
					"</a>",
				),
			},
			expectedStr: joinLines(
				xmlstruct.DefaultHeader,
				``,
				`package main`,
				``,
				`import "encoding/xml"`,
				``,
				`type A struct {`,
				"\tXMLName xml.Name `xml:\"a\"`",
				"\tB       int64    `xml:\"b\"`",
				`}`,
			),
		},
		{
			name: "string_attribute",
			options: []xmlstruct.GeneratorOption{
				xmlstruct.WithIntType("int64"),
			},
			xmlStrs: []string{
				joinLines(
					"<a>",
					`  <b id=""/>`,
					`  <b id="c"/>`,
					"</a>",
				),
			},
			expectedStr: joinLines(
				xmlstruct.DefaultHeader,
				``,
				`package main`,
				``,
				`import "encoding/xml"`,
				``,
				`type A struct {`,
				"\tXMLName xml.Name `xml:\"a\"`",
				"\tB       []struct {",
				"\t\tId string `xml:\"id,attr\"`",
				"\t} `xml:\"b\"`",
				`}`,
			),
		},
		{
			name: "empty_attribute",
			options: []xmlstruct.GeneratorOption{
				xmlstruct.WithIntType("int64"),
			},
			xmlStrs: []string{
				joinLines(
					"<a>",
					`  <b id=""/>`,
					"</a>",
				),
			},
			expectedStr: joinLines(
				xmlstruct.DefaultHeader,
				``,
				`package main`,
				``,
				`import "encoding/xml"`,
				``,
				`type A struct {`,
				"\tXMLName xml.Name `xml:\"a\"`",
				"\tB       struct {",
				"\t\tId string `xml:\"id,attr\"`",
				"\t} `xml:\"b\"`",
				`}`,
			),
		},
		{
			name: "empty_struct",
			xmlStrs: []string{
				joinLines(
					"<a>",
					"  <b/>",
					"</a>",
				),
			},
			expectedStr: joinLines(
				xmlstruct.DefaultHeader,
				``,
				`package main`,
				``,
				`import "encoding/xml"`,
				``,
				`type A struct {`,
				"\tXMLName xml.Name `xml:\"a\"`",
				"\tB       struct{} `xml:\"b\"`",
				`}`,
			),
		},
		{
			name: "empty_structs",
			xmlStrs: []string{
				joinLines(
					"<a>",
					"  <b/>",
					"  <b/>",
					"</a>",
				),
			},
			expectedStr: joinLines(
				xmlstruct.DefaultHeader,
				``,
				`package main`,
				``,
				`import "encoding/xml"`,
				``,
				`type A struct {`,
				"\tXMLName xml.Name   `xml:\"a\"`",
				"\tB       []struct{} `xml:\"b\"`",
				`}`,
			),
		},
		{
			name: "empty_top_level_type",
			xmlStrs: []string{
				"<a/>",
			},
			expectedStr: joinLines(
				xmlstruct.DefaultHeader,
				``,
				`package main`,
				``,
				`import "encoding/xml"`,
				``,
				`type A struct {`,
				"\tXMLName xml.Name `xml:\"a\"`",
				`}`,
			),
		},
		{
			name: "multiple_top_level_types",
			xmlStrs: []string{
				"<c/>",
				"<b/>",
				"<a/>",
			},
			expectedStr: joinLines(
				xmlstruct.DefaultHeader,
				``,
				`package main`,
				``,
				`import "encoding/xml"`,
				``,
				`type A struct {`,
				"\tXMLName xml.Name `xml:\"a\"`",
				`}`,
				``,
				`type B struct {`,
				"\tXMLName xml.Name `xml:\"b\"`",
				`}`,
				``,
				`type C struct {`,
				"\tXMLName xml.Name `xml:\"c\"`",
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
