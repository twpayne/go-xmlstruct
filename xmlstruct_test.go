package xmlstruct

import (
	"encoding/xml"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultExportNameFunc(t *testing.T) {
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
		t.Run(tc.localName, func(t *testing.T) {
			xmlName := xml.Name{
				Local: tc.localName,
			}
			assert.Equal(t, tc.expected, DefaultExportNameFunc(xmlName))
		})
	}
}
