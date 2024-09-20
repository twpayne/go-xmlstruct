package xmlstruct_test

import (
	"encoding/xml"
	"testing"

	"github.com/alecthomas/assert/v2"

	"github.com/twpayne/go-xmlstruct"
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
		{
			localName: "+",
			expected:  "_",
		},
	} {
		t.Run(tc.localName, func(t *testing.T) {
			t.Parallel()

			xmlName := xml.Name{
				Local: tc.localName,
			}
			assert.Equal(t, tc.expected, xmlstruct.DefaultExportNameFunc(xmlName))
		})
	}
}

func TestDefaultUnexportNameFunc(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		localName string
		expected  string
	}{
		{
			localName: "id",
			expected:  "id",
		},
		{
			localName: "ID",
			expected:  "id",
		},
		{
			localName: "Id",
			expected:  "id",
		},
		{
			localName: "camelCase",
			expected:  "camelCase",
		},
		{
			localName: "camelCaseId",
			expected:  "camelCaseID",
		},
		{
			localName: "kebab-case",
			expected:  "kebabCase",
		},
		{
			localName: "kebab--case",
			expected:  "kebabCase",
		},
		{
			localName: "kebab-id",
			expected:  "kebabID",
		},
		{
			localName: "snake_case",
			expected:  "snakeCase",
		},
		{
			localName: "snake__case",
			expected:  "snakeCase",
		},
		{
			localName: "snake-id",
			expected:  "snakeID",
		},
		{
			localName: "+",
			expected:  "_",
		},
	} {
		t.Run(tc.localName, func(t *testing.T) {
			t.Parallel()

			xmlName := xml.Name{
				Local: tc.localName,
			}
			assert.Equal(t, tc.expected, xmlstruct.DefaultUnexportNameFunc(xmlName))
		})
	}
}
