package xmlstruct_test

import (
	"strings"
	"testing"

	"github.com/alecthomas/assert/v2"

	"github.com/twpayne/go-xmlstruct"
)

func TestGenerator(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name        string
		xmlStr      string
		xmlStrs     []string
		options     []xmlstruct.GeneratorOption
		expectedStr string
		expectedErr string
	}{
		{
			name: "simple_string",
			xmlStr: joinLines(
				"<a>",
				"  <b>c</b>",
				"</a>",
			),
			expectedStr: joinLines(
				xmlstruct.DefaultHeader,
				``,
				`package main`,
				``,
				`type A struct {`,
				"\tB string `xml:\"b\"`",
				`}`,
			),
		},
		{
			name: "simple_int",
			xmlStr: joinLines(
				"<a>",
				"  <b>2</b>",
				"</a>",
			),
			expectedStr: joinLines(
				xmlstruct.DefaultHeader,
				``,
				`package main`,
				``,
				`type A struct {`,
				"\tB int `xml:\"b\"`",
				`}`,
			),
		},
		{
			name: "int_type",
			options: []xmlstruct.GeneratorOption{
				xmlstruct.WithHeader("// Custom header."),
			},
			xmlStr: joinLines(
				"<a>",
				"  <b>c</b>",
				"</a>",
			),
			expectedStr: joinLines(
				`// Custom header.`,
				``,
				`package main`,
				``,
				`type A struct {`,
				"\tB string `xml:\"b\"`",
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
				`type A struct {`,
				"\tB int64 `xml:\"b\"`",
				`}`,
			),
		},
		{
			name: "string_attribute",
			options: []xmlstruct.GeneratorOption{
				xmlstruct.WithIntType("int64"),
			},
			xmlStr: joinLines(
				"<a>",
				`  <b id=""/>`,
				`  <b id="c"/>`,
				"</a>",
			),
			expectedStr: joinLines(
				xmlstruct.DefaultHeader,
				``,
				`package main`,
				``,
				`type A struct {`,
				"\tB []struct {",
				"\t\tID string `xml:\"id,attr\"`",
				"\t} `xml:\"b\"`",
				`}`,
			),
		},
		{
			name: "empty_attribute",
			options: []xmlstruct.GeneratorOption{
				xmlstruct.WithIntType("int64"),
			},
			xmlStr: joinLines(
				"<a>",
				`  <b id=""/>`,
				"</a>",
			),
			expectedStr: joinLines(
				xmlstruct.DefaultHeader,
				``,
				`package main`,
				``,
				`type A struct {`,
				"\tB struct {",
				"\t\tID string `xml:\"id,attr\"`",
				"\t} `xml:\"b\"`",
				`}`,
			),
		},
		{
			name: "empty_struct",
			xmlStr: joinLines(
				"<a>",
				"  <b/>",
				"</a>",
			),
			expectedStr: joinLines(
				xmlstruct.DefaultHeader,
				``,
				`package main`,
				``,
				`type A struct {`,
				"\tB struct{} `xml:\"b\"`",
				`}`,
			),
		},
		{
			name: "empty_structs",
			xmlStr: joinLines(
				"<a>",
				"  <b/>",
				"  <b/>",
				"</a>",
			),
			expectedStr: joinLines(
				xmlstruct.DefaultHeader,
				``,
				`package main`,
				``,
				`type A struct {`,
				"\tB []struct{} `xml:\"b\"`",
				`}`,
			),
		},
		{
			name:   "empty_top_level_type",
			xmlStr: "<a/>",
			expectedStr: joinLines(
				xmlstruct.DefaultHeader,
				``,
				`package main`,
				``,
				`type A struct{}`,
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
				`type A struct{}`,
				``,
				`type B struct{}`,
				``,
				`type C struct{}`,
			),
		},
		{
			name: "named_types",
			options: []xmlstruct.GeneratorOption{
				xmlstruct.WithNamedTypes(true),
			},
			xmlStr: joinLines(
				`<a>`,
				`  <b>`,
				`    <c/>`,
				`  </b>`,
				`</a>`,
			),
			expectedStr: joinLines(
				xmlstruct.DefaultHeader,
				``,
				`package main`,
				``,
				`type A struct {`,
				"\tB B `xml:\"b\"`", //nolint:dupword
				`}`,
				``,
				`type B struct {`,
				"\tC struct{} `xml:\"c\"`",
				`}`,
			),
		},
		// FIXME make the following test pass
		/*
			{
				name:    "duplicate_field_name",
				options: []xmlstruct.GeneratorOption{},
				xmlStrs: []string{
					joinLines(
						`<a>`,
						`  <b/>`,
						`  <B/>`,
						`</a>`,
					),
				},
				expectedErr: "B: duplicate field name",
			},
		*/
		{
			name: "duplicate_type_name",
			options: []xmlstruct.GeneratorOption{
				xmlstruct.WithNamedTypes(true),
			},
			xmlStr: joinLines(
				`<a>`,
				`  <b>`,
				`    <c/>`,
				`  </b>`,
				`  <B>`,
				`    <c/>`,
				`  </B>`,
				`</a>`,
			),
			expectedErr: "B: duplicate type name",
		},
		{
			name: "with_top_level_attributes",
			options: []xmlstruct.GeneratorOption{
				xmlstruct.WithTopLevelAttributes(true),
			},
			xmlStr: `<a b="0"/>`,
			expectedStr: joinLines(
				xmlstruct.DefaultHeader,
				``,
				`package main`,
				``,
				`type A struct {`,
				"\tB bool `xml:\"b,attr\"`",
				`}`,
			),
		},
		{
			name: "without_top_level_attributes",
			options: []xmlstruct.GeneratorOption{
				xmlstruct.WithTopLevelAttributes(false),
			},
			xmlStr: `<a b="0"/>`,
			expectedStr: joinLines(
				xmlstruct.DefaultHeader,
				``,
				`package main`,
				``,
				`type A struct{}`,
			),
		},
		{
			name: "char_data_field_name",
			options: []xmlstruct.GeneratorOption{
				xmlstruct.WithCharDataFieldName("Text"),
			},
			xmlStr: joinLines(
				`<a>`,
				`  <b id="c">d</b>`,
				`</a>`,
			),
			expectedStr: joinLines(
				xmlstruct.DefaultHeader,
				``,
				`package main`,
				``,
				`type A struct {`,
				"\tB struct {",
				"\t\tID   string `xml:\"id,attr\"`",
				"\t\tText string `xml:\",chardata\"`",
				"\t} `xml:\"b\"`",
				`}`,
			),
		},
		{
			name:    "test_int_parse",
			options: []xmlstruct.GeneratorOption{},
			xmlStr: joinLines(
				`<a>`,
				`  <b index="0">one</b>`,
				`  <b index="1">two</b>`,
				`  <b index="2">three</b>`,
				`</a>`,
			),
			expectedStr: joinLines(
				xmlstruct.DefaultHeader,
				``,
				`package main`,
				``,
				`type A struct {`,
				"\tB []struct {",
				"\t\tIndex    int    `xml:\"index,attr\"`",
				"\t\tCharData string `xml:\",chardata\"`",
				"\t} `xml:\"b\"`",
				`}`,
			),
		},
		{
			name:    "test_int_parse_without_1_0",
			options: []xmlstruct.GeneratorOption{},
			xmlStr: joinLines(
				`<a>`,
				`  <b index="2">two</b>`,
				`  <b index="3">three</b>`,
				`  <b index="4">four</b>`,
				`</a>`,
			),
			expectedStr: joinLines(
				xmlstruct.DefaultHeader,
				``,
				`package main`,
				``,
				`type A struct {`,
				"\tB []struct {",
				"\t\tIndex    int    `xml:\"index,attr\"`",
				"\t\tCharData string `xml:\",chardata\"`",
				"\t} `xml:\"b\"`",
				`}`,
			),
		},
		{
			name:    "test_int_parse_data",
			options: []xmlstruct.GeneratorOption{},
			xmlStr: joinLines(
				`<a>`,
				`  <b>0</b>`,
				`  <b>1</b>`,
				`  <b>2</b>`,
				`</a>`,
			),
			expectedStr: joinLines(
				xmlstruct.DefaultHeader,
				``,
				`package main`,
				``,
				`type A struct {`,
				"\tB []int `xml:\"b\"`",
				`}`,
			),
		},
		{
			name:    "test_int_parse_data_without_1_0",
			options: []xmlstruct.GeneratorOption{},
			xmlStr: joinLines(
				`<a>`,
				`  <b>2</b>`,
				`  <b>3</b>`,
				`  <b>4</b>`,
				`</a>`,
			),
			expectedStr: joinLines(
				xmlstruct.DefaultHeader,
				``,
				`package main`,
				``,
				`type A struct {`,
				"\tB []int `xml:\"b\"`",
				`}`,
			),
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			generator := xmlstruct.NewGenerator(tc.options...)
			if tc.xmlStr != "" {
				assert.NoError(t, generator.ObserveReader(strings.NewReader(tc.xmlStr)))
			}
			for _, xmlStr := range tc.xmlStrs {
				assert.NoError(t, generator.ObserveReader(strings.NewReader(xmlStr)))
			}
			actual, err := generator.Generate()
			if tc.expectedErr != "" {
				assert.EqualError(t, err, tc.expectedErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedStr, string(actual))
			}
		})
	}
}

func joinLines(lines ...string) string {
	return strings.Join(lines, "\n") + "\n"
}
