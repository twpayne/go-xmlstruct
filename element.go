package xmlstruct

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"

	"golang.org/x/exp/maps"
)

// An element describes an observed XML element, its attributes, chardata, and
// children.
type element struct {
	attrValues    map[xml.Name]*value
	charDataValue value
	childElements map[xml.Name]*element
	name          xml.Name
	optional      bool
	repeated      bool
}

// newElement returns a new element.
func newElement(name xml.Name) *element {
	return &element{
		name:          name,
		attrValues:    make(map[xml.Name]*value),
		childElements: make(map[xml.Name]*element),
	}
}

// observeAttrs updates e's observed attributes with attrs.
func (e *element) observeAttrs(attrs []xml.Attr, options *observeOptions) {
	attrCounts := make(map[xml.Name]int)
	for _, attr := range attrs {
		attrName := options.nameFunc(attr.Name)
		attrCounts[attrName]++
		attrValue, ok := e.attrValues[attrName]
		if !ok {
			attrValue = &value{}
			e.attrValues[attrName] = attrValue
		}
		attrValue.observe(attr.Value, options)
	}
	for attrName, count := range attrCounts {
		if count > 1 {
			e.attrValues[attrName].repeated = true
		}
	}
	for attrName, attrValue := range e.attrValues {
		if attrCounts[attrName] == 0 {
			attrValue.optional = true
		}
	}
}

// observeChildElement updates e's observed chardata and child elements with
// tokens read from decoder.
func (e *element) observeChildElement(decoder *xml.Decoder, startElement xml.StartElement, options *observeOptions) error {
	e.observeAttrs(startElement.Attr, options)
	childCounts := make(map[xml.Name]int)
FOR:
	for {
		token, err := decoder.Token()
		if err != nil {
			return err
		}
		switch token := token.(type) {
		case xml.StartElement:
			childName := options.nameFunc(token.Name)
			childCounts[childName]++
			childElement, ok := e.childElements[childName]
			if !ok {
				childElement = newElement(childName)
				e.childElements[childName] = childElement
			}
			if err := childElement.observeChildElement(decoder, token, options); err != nil {
				return err
			}
		case xml.EndElement:
			break FOR
		case xml.CharData:
			if trimmedToken := bytes.TrimSpace(token); len(trimmedToken) > 0 {
				e.charDataValue.observe(string(token), options)
			}
		}
	}
	for childName, count := range childCounts {
		if count > 1 {
			e.childElements[childName].repeated = true
		}
	}
	for childName, childElement := range e.childElements {
		if childCounts[childName] == 0 {
			childElement.optional = true
		}
	}
	return nil
}

// writeGoType writes e's Go type to w.
func (e *element) writeGoType(w io.Writer, options *generateOptions, indentPrefix string) {
	prefix := ""
	if e.repeated {
		prefix += "[]"
	}
	if options.usePointersForOptionalFields && e.optional {
		prefix += "*"
	}

	if indentPrefix != "" && len(e.attrValues) == 0 && len(e.childElements) == 0 {
		fmt.Fprintf(w, "%s%s", prefix, e.charDataValue.goType(options))
		return
	}

	fmt.Fprintf(w, "%sstruct {\n", prefix)

	if indentPrefix == "" {
		options.importPackageNames["encoding/xml"] = struct{}{}
		fmt.Fprintf(w, "%s\tXMLName xml.Name `xml:\"%s\"`\n", indentPrefix, e.name.Local)
	}

	for _, attrName := range sortXMLNames(maps.Keys(e.attrValues)) {
		attrValue := e.attrValues[attrName]
		fmt.Fprintf(w, "%s\t%s %s `xml:\"%s,attr\"`\n", indentPrefix, options.exportNameFunc(attrName), attrValue.goType(options), attrName.Local)
	}

	if e.charDataValue.observations > 0 {
		fmt.Fprintf(w, "%s\tCharData string `xml:\",chardata\"`\n", indentPrefix)
	}

	for _, childName := range sortXMLNames(maps.Keys(e.childElements)) {
		childElement := e.childElements[childName]
		fmt.Fprintf(w, "%s\t%s ", indentPrefix, options.exportNameFunc(childName))
		childElement.writeGoType(w, options, indentPrefix+"\t")
		fmt.Fprintf(w, " `xml:\"%s\"`\n", childName.Local)
	}

	fmt.Fprintf(w, "%s}", indentPrefix)
}
