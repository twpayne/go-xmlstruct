package xmlstruct

import (
	"encoding/xml"
	"unicode"

	"golang.org/x/exp/slices"
)

type ExportNameFunc func(xml.Name) string

type NameFunc func(xml.Name) xml.Name

type observeOptions struct {
	nameFunc   NameFunc
	timeLayout string
}

type sourceOptions struct {
	exportNameFunc               ExportNameFunc
	importPackageNames           map[string]struct{}
	usePointersForOptionalFields bool
}

func DefaultExportNameFunc(name xml.Name) string {
	runes := []rune(name.Local)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

func IdentityNameFunc(name xml.Name) xml.Name {
	return name
}

func IgnoreNamespaceNameFunc(name xml.Name) xml.Name {
	return xml.Name{
		Local: name.Local,
	}
}

func sortXMLNames(xmlNames []xml.Name) []xml.Name {
	slices.SortFunc(xmlNames, func(a, b xml.Name) bool {
		switch {
		case a.Space < b.Space:
			return true
		case a.Space == b.Space:
			return a.Local < b.Local
		default:
			return false
		}
	})
	return xmlNames
}
