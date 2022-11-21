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
		expectedErr string
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
				`type A struct {`,
				"\tB string `xml:\"b\"`",
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
				`type A struct {`,
				"\tB []struct {",
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
				`type A struct {`,
				"\tB struct {",
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
				`type A struct {`,
				"\tB struct{} `xml:\"b\"`",
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
				`type A struct {`,
				"\tB []struct{} `xml:\"b\"`",
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
			xmlStrs: []string{
				joinLines(
					`<a>`,
					`  <b>`,
					`    <c/>`,
					`  </b>`,
					`</a>`,
				),
			},
			expectedStr: joinLines(
				xmlstruct.DefaultHeader,
				``,
				`package main`,
				``,
				`type A struct {`,
				"\tB B `xml:\"b\"`",
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
			xmlStrs: []string{
				joinLines(
					`<a>`,
					`  <b>`,
					`    <c/>`,
					`  </b>`,
					`  <B>`,
					`    <c/>`,
					`  </B>`,
					`</a>`,
				),
			},
			expectedErr: "B: duplicate type name",
		},
		{
			name: "with_top_level_attributes",
			options: []xmlstruct.GeneratorOption{
				xmlstruct.WithTopLevelAttributes(true),
			},
			xmlStrs: []string{
				`<a b="0"/>`,
			},
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
			xmlStrs: []string{
				`<a b="0"/>`,
			},
			expectedStr: joinLines(
				xmlstruct.DefaultHeader,
				``,
				`package main`,
				``,
				`type A struct{}`,
			),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			generator := xmlstruct.NewGenerator(tc.options...)
			for _, xmlStr := range tc.xmlStrs {
				require.NoError(t, generator.ObserveReader(strings.NewReader(xmlStr)))
			}
			actual, err := generator.Generate()
			if tc.expectedErr != "" {
				require.EqualError(t, err, tc.expectedErr)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expectedStr, string(actual))
			}
		})
	}
}

func joinLines(lines ...string) string {
	return strings.Join(lines, "\n") + "\n"
}
