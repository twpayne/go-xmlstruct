package xmlstruct

import (
	"encoding/xml"
	"testing"

	"github.com/alecthomas/assert/v2"
)

func TestDefaultExportNameFunc(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		localName string
		expected  string
	}{
		{
			localName: "id",
			expected:  "ID",
		},
		{
			localName: "camelCase",
			expected:  "CamelCase",
		},
		{
			localName: "camelCaseId",
			expected:  "CamelCaseID",
		},
		{
			localName: "kebab-case",
			expected:  "KebabCase",
		},
		{
			localName: "kebab--case",
			expected:  "KebabCase",
		},
		{
			localName: "kebab-id",
			expected:  "KebabID",
		},
		{
			localName: "snake_case",
			expected:  "SnakeCase",
		},
		{
			localName: "snake__case",
			expected:  "SnakeCase",
		},
		{
			localName: "snake-id",
			expected:  "SnakeID",
		},
	} {
		tc := tc
		t.Run(tc.localName, func(t *testing.T) {
			t.Parallel()

			xmlName := xml.Name{
				Local: tc.localName,
			}
			assert.Equal(t, tc.expected, DefaultExportNameFunc(xmlName))
		})
	}
}
