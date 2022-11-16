package xmlstruct

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"

	"golang.org/x/exp/maps"
)

type observedElement struct {
	name     xml.Name
	repeated bool
	optional bool
	attrs    map[xml.Name]*ObservedValue
	children map[xml.Name]*observedElement
	charData ObservedValue
}

func newObservedElement(name xml.Name) *observedElement {
	return &observedElement{
		name:     name,
		attrs:    make(map[xml.Name]*ObservedValue),
		children: make(map[xml.Name]*observedElement),
	}
}

func (e *observedElement) observeAttrs(attrs []xml.Attr, options *observeOptions) {
	attrCounts := make(map[xml.Name]int)
	for _, attr := range attrs {
		attrName := options.nameFunc(attr.Name)
		attrCounts[attrName]++
		observedValue, ok := e.attrs[attrName]
		if !ok {
			observedValue = &ObservedValue{}
			e.attrs[attrName] = observedValue
		}
		observedValue.observe(attr.Value, options)
	}
	for attrName, count := range attrCounts {
		if count > 1 {
			e.attrs[attrName].repeated = true
		}
	}
	for attrName, observedValue := range e.attrs {
		if attrCounts[attrName] == 0 {
			observedValue.optional = true
		}
	}
}

func (e *observedElement) observeChildElement(decoder *xml.Decoder, startElement xml.StartElement, options *observeOptions) error {
	childCounts := make(map[xml.Name]int)
	e.observeAttrs(startElement.Attr, options)
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
			child, ok := e.children[childName]
			if !ok {
				child = newObservedElement(childName)
				e.children[childName] = child
			}
			if err := child.observeChildElement(decoder, token, options); err != nil {
				return err
			}
		case xml.EndElement:
			break FOR
		case xml.CharData:
			if trimmedToken := bytes.TrimSpace(token); len(trimmedToken) > 0 {
				e.charData.observe(string(token), options)
			}
		}
	}
	for childName, count := range childCounts {
		if count > 1 {
			e.children[childName].repeated = true
		}
	}
	for childName, observedElement := range e.children {
		if childCounts[childName] == 0 {
			observedElement.optional = true
		}
	}
	return nil
}

func (e *observedElement) writeGoType(w io.Writer, options *sourceOptions, indentPrefix string) {
	prefix := ""
	if e.repeated {
		prefix += "[]"
	}
	if options.usePointersForOptionalFields && e.optional {
		prefix += "*"
	}

	if len(e.attrs) == 0 && len(e.children) == 0 {
		fmt.Fprintf(w, "%s%s", prefix, e.charData.goType(options))
		return
	}

	fmt.Fprintf(w, "%sstruct {\n", prefix)

	if indentPrefix == "" {
		options.importPackageNames["encoding/xml"] = struct{}{}
		fmt.Fprintf(w, "%s\tXMLName xml.Name `xml:\"%s\"`\n", indentPrefix, e.name.Local)
	}

	if e.charData.observations > 0 {
		fmt.Fprintf(w, "%s\tCharData string `xml:\",chardata\"`\n", indentPrefix)
	}

	for _, attrName := range sortXMLNames(maps.Keys(e.attrs)) {
		observedValue := e.attrs[attrName]
		fmt.Fprintf(w, "%s\t%s %s `xml:\"%s,attr\"`\n", indentPrefix, options.exportNameFunc(attrName), observedValue.goType(options), attrName.Local)
	}

	for _, childName := range sortXMLNames(maps.Keys(e.children)) {
		observedElement := e.children[childName]
		fmt.Fprintf(w, "%s\t%s ", indentPrefix, options.exportNameFunc(childName))
		observedElement.writeGoType(w, options, indentPrefix+"\t")
		fmt.Fprintf(w, " `xml:\"%s\"`\n", childName.Local)
	}

	fmt.Fprintf(w, "%s}", indentPrefix)
}
