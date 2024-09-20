// Package xmlstruct generates Go structs from multiple XML documents.
package xmlstruct

import (
	"cmp"
	"encoding/xml"
	"regexp"
	"slices"
	"strings"
	"unicode"
)

const (
	DefaultAttrNameSuffix               = ""
	DefaultCharDataFieldName            = "CharData"
	DefaultElemNameSuffix               = ""
	DefaultFormatSource                 = true
	DefaultHeader                       = "// Code generated by goxmlstruct. DO NOT EDIT."
	DefaultTopLevelAttributes           = false
	DefaultImports                      = true
	DefaultIntType                      = "int"
	DefaultNamedRoot                    = false
	DefaultNamedTypes                   = false
	DefaultCompactTypes                 = false
	DefaultPackageName                  = "main"
	DefaultPreserveOrder                = false
	DefaultTimeLayout                   = "2006-01-02T15:04:05Z"
	DefaultUsePointersForOptionalFields = true
	DefaultUseRawToken                  = false
	DefaultEmptyElements                = true
)

var (
	// TitleFirstRuneExportNameFunc returns name.Local with the initial rune
	// capitalized.
	TitleFirstRuneExportNameFunc = func(name xml.Name) string {
		runes := []rune(name.Local)
		runes[0] = unicode.ToUpper(runes[0])
		return string(runes)
	}

	kebabOrSnakeCaseWordBoundaryRx = regexp.MustCompile(`[-_]+\pL`)
	nonIdentifierRuneRx            = regexp.MustCompile(`[^\pL\pN]`)

	// DefaultExportNameFunc returns name.Local with kebab- and snakecase words
	// converted to UpperCamelCase and any Id suffix converted to ID.
	DefaultExportNameFunc = func(name xml.Name) string {
		localName := kebabOrSnakeCaseWordBoundaryRx.ReplaceAllStringFunc(name.Local, func(s string) string {
			return strings.ToUpper(s[len(s)-1:])
		})
		localName = nonIdentifierRuneRx.ReplaceAllLiteralString(localName, "_")
		runes := []rune(localName)
		runes[0] = unicode.ToUpper(runes[0])
		if len(runes) > 1 && runes[len(runes)-2] == 'I' && runes[len(runes)-1] == 'd' {
			runes[len(runes)-1] = 'D'
		}
		return string(runes)
	}

	// DefaultUnexportNameFunc returns name.Local with kebab- and snakecase words
	// converted to lowerCamelCase
	// Any ID prefix is converted to id, and any Id suffix converted to ID.
	DefaultUnexportNameFunc = func(name xml.Name) string {
		localName := kebabOrSnakeCaseWordBoundaryRx.ReplaceAllStringFunc(name.Local, func(s string) string {
			return strings.ToUpper(s[len(s)-1:])
		})
		localName = nonIdentifierRuneRx.ReplaceAllLiteralString(localName, "_")
		runes := []rune(localName)
		runes[0] = unicode.ToLower(runes[0])
		if len(runes) > 1 {
			if runes[len(runes)-2] == 'I' && runes[len(runes)-1] == 'd' {
				runes[len(runes)-1] = 'D'
			}
			if runes[0] == 'i' && runes[1] == 'D' {
				runes[1] = 'd'
			}
		}
		return string(runes)
	}
)

var (
	// IgnoreNamespaceNameFunc returns name with name.Space cleared. The same
	// local name in different namespaces will be treated as identical names.
	IgnoreNamespaceNameFunc = func(name xml.Name) xml.Name {
		return xml.Name{
			Local: name.Local,
		}
	}

	// The IdentityNameFunc returns name unchanged. The same local name in
	// different namespaces will be treated as distinct names.
	IdentityNameFunc = func(name xml.Name) xml.Name {
		return name
	}

	DefaultNameFunc = IgnoreNamespaceNameFunc
)

// An ExportNameFunc returns the exported Go identifier for the given xml.Name.
type ExportNameFunc func(xml.Name) string

// A NameFunc modifies xml.Names observed in the XML documents.
type NameFunc func(xml.Name) xml.Name

// observeOptions contains options for observing XML documents.
type observeOptions struct {
	getOrder           func() int
	nameFunc           NameFunc
	timeLayout         string
	typeOrder          map[xml.Name]int
	topLevelAttributes bool
	topLevelElements   map[xml.Name]*element
	useRawToken        bool
}

// generateOptions contains options for generating Go source.
type generateOptions struct {
	attrNameSuffix               string
	charDataFieldName            string
	elemNameSuffix               string
	exportNameFunc               ExportNameFunc
	exportTypeNameFunc           ExportNameFunc
	header                       string
	importPackageNames           map[string]struct{}
	intType                      string
	namedRoot                    bool
	namedTypes                   map[xml.Name]*element
	compactTypes                 bool
	preserveOrder                bool
	simpleTypes                  map[xml.Name]struct{}
	usePointersForOptionalFields bool
	emptyElements                bool
}

func mapKeys[M ~map[K]V, K comparable, V any](m M) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func mapValues[M ~map[K]V, K comparable, V any](m M) []V {
	values := make([]V, 0, len(m))
	for _, v := range m {
		values = append(values, v)
	}
	return values
}

// sortedKeys returns the keys of m in order.
func sortedKeys[M ~map[K]V, K cmp.Ordered, V any](m M) []K {
	keys := mapKeys(m)
	slices.Sort(keys)
	return keys
}
