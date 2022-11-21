package xmlstruct

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
)

// An element describes an observed XML element, its attributes, chardata, and
// children.
type element struct {
	attrValues       map[xml.Name]*value
	charDataValue    value
	childElements    map[xml.Name]*element
	name             xml.Name
	optionalChildren map[xml.Name]struct{}
	repeatedChildren map[xml.Name]struct{}
}

// newElement returns a new element.
func newElement(name xml.Name) *element {
	return &element{
		name:             name,
		attrValues:       make(map[xml.Name]*value),
		childElements:    make(map[xml.Name]*element),
		optionalChildren: make(map[xml.Name]struct{}),
		repeatedChildren: make(map[xml.Name]struct{}),
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
			attrValue = &value{
				name: attrName,
			}
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
func (e *element) observeChildElement(decoder *xml.Decoder, startElement xml.StartElement, depth int, options *observeOptions) error {
	if options.topLevelAttributes || depth != 0 {
		e.observeAttrs(startElement.Attr, options)
	}
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
				if options.topLevelElements != nil {
					if topLevelElement, ok := options.topLevelElements[childName]; ok {
						childElement = topLevelElement
					} else {
						topLevelElement = newElement(childName)
						options.topLevelElements[childName] = topLevelElement
						childElement = topLevelElement
					}
				} else {
					childElement = newElement(childName)
				}
				e.childElements[childName] = childElement
			}
			if err := childElement.observeChildElement(decoder, token, depth+1, options); err != nil {
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
			e.repeatedChildren[childName] = struct{}{}
		}
	}
	for childName := range e.childElements {
		if childCounts[childName] == 0 {
			e.optionalChildren[childName] = struct{}{}
		}
	}
	return nil
}

// writeGoType writes e's Go type to w.
func (e *element) writeGoType(w io.Writer, options *generateOptions, indentPrefix string) error {
	if len(e.attrValues) == 0 && len(e.childElements) == 0 {
		fmt.Fprintf(w, "%s", e.charDataValue.goType(options))
		return nil
	}

	fmt.Fprintf(w, "struct {\n")

	fieldNames := make(map[string]struct{})

	attrValuesByExportedName := make(map[string]*value, len(e.attrValues))
	for attrName, attrValue := range e.attrValues {
		exportedAttrName := options.exportNameFunc(attrName)
		if _, ok := fieldNames[exportedAttrName]; ok {
			return fmt.Errorf("%s: duplicate field name", exportedAttrName)
		}
		fieldNames[exportedAttrName] = struct{}{}
		attrValuesByExportedName[exportedAttrName] = attrValue
	}
	for _, exportedAttrName := range sortedKeys(attrValuesByExportedName) {
		attrValue := attrValuesByExportedName[exportedAttrName]
		fmt.Fprintf(w, "%s\t%s %s `xml:\"%s,attr\"`\n", indentPrefix, exportedAttrName, attrValue.goType(options), attrValue.name.Local)
	}

	if e.charDataValue.observations > 0 {
		fieldName := "CharData"
		if _, ok := fieldNames[fieldName]; ok {
			return fmt.Errorf("%s: duplicate field name", fieldName)
		}
		fieldNames[fieldName] = struct{}{}
		fmt.Fprintf(w, "%s\t%s string `xml:\",chardata\"`\n", indentPrefix, fieldName)
	}

	childElementsByExportedName := make(map[string]*element, len(e.childElements))
	for childName, childElement := range e.childElements {
		exportedChildName := options.exportNameFunc(childName)
		childElementsByExportedName[exportedChildName] = childElement
	}
	for _, exportedChildName := range sortedKeys(childElementsByExportedName) {
		if _, ok := fieldNames[exportedChildName]; ok {
			fieldNames[exportedChildName] = struct{}{}
		}
		fieldNames[exportedChildName] = struct{}{}

		childElement := childElementsByExportedName[exportedChildName]
		fmt.Fprintf(w, "%s\t%s ", indentPrefix, exportedChildName)
		if _, repeated := e.repeatedChildren[childElement.name]; repeated {
			fmt.Fprintf(w, "[]")
		}
		if options.usePointersForOptionalFields {
			if _, optional := e.optionalChildren[childElement.name]; optional {
				fmt.Fprintf(w, "*")
			}
		}
		if topLevelElement, ok := options.namedTypes[childElement.name]; ok {
			fmt.Fprintf(w, "%s", options.exportNameFunc(topLevelElement.name))
		} else if _, ok := options.simpleTypes[childElement.name]; ok {
			fmt.Fprintf(w, "%s", childElement.charDataValue.goType(options))
		} else {
			if err := childElement.writeGoType(w, options, indentPrefix+"\t"); err != nil {
				return err
			}
		}
		fmt.Fprintf(w, " `xml:\"%s\"`\n", childElement.name.Local)
	}

	fmt.Fprintf(w, "%s}", indentPrefix)
	return nil
}
