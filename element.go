package xmlstruct

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"maps"
	"slices"
	"strings"
)

// An element describes an observed XML element, its attributes, chardata, and
// children.
type element struct {
	attrValues       map[xml.Name]*value
	charDataValue    value
	childElements    map[xml.Name]*element
	nestedCount      int
	childOrder       map[xml.Name]int
	name             xml.Name
	optionalChildren map[xml.Name]struct{}
	repeatedChildren map[xml.Name]struct{}
	root             bool
}

// newElement returns a new element.
func newElement(name xml.Name) *element {
	return &element{
		name:             name,
		attrValues:       make(map[xml.Name]*value),
		childElements:    make(map[xml.Name]*element),
		childOrder:       make(map[xml.Name]int),
		optionalChildren: make(map[xml.Name]struct{}),
		repeatedChildren: make(map[xml.Name]struct{}),
	}
}

// observeAttrs updates e's observed attributes with attrs.
func (e *element) observeAttrs(attrs []xml.Attr, options *observeOptions) {
	attrCounts := make(map[xml.Name]int)
	for _, attr := range attrs {
		attrName := options.nameFunc(attr.Name)
		if attrName == (xml.Name{}) {
			continue
		}
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
		var token xml.Token
		var err error
		if options.useRawToken {
			token, err = decoder.RawToken()
		} else {
			token, err = decoder.Token()
		}
		if err != nil {
			return err
		}
		switch token := token.(type) {
		case xml.StartElement:
			childName := options.nameFunc(token.Name)
			if childName == (xml.Name{}) {
				break
			}
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
					if _, ok := options.typeOrder[childName]; !ok {
						options.typeOrder[childName] = options.getOrder()
					}
				} else {
					childElement = newElement(childName)
				}
				e.childElements[childName] = childElement
			}
			if childElement == e {
				e.nestedCount++
			}
			if _, ok := e.childOrder[childName]; !ok {
				e.childOrder[childName] = options.getOrder()
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
	if options.compactTypes && e.isContainer() {
		for _, v := range e.childElements {
			if v == e {
				fmt.Fprintf(w, "%s", e.charDataValue.goType(options))
				return nil
			}
		}
	}

	if len(e.attrValues) == 0 && len(e.childElements) == 0 && (!e.root || !options.namedRoot) {
		fmt.Fprintf(w, "%s", e.charDataValue.goType(options))
		return nil
	}

	fmt.Fprintf(w, "struct {\n")

	fieldNames := make(map[string]struct{})

	attrValuesByExportedName := make(map[string]*value, len(e.attrValues))
	for attrName, attrValue := range e.attrValues {
		exportedAttrName := options.exportNameFunc(attrName) + options.attrNameSuffix
		if _, ok := fieldNames[exportedAttrName]; ok {
			return fmt.Errorf("%s: duplicate field name", exportedAttrName)
		}
		fieldNames[exportedAttrName] = struct{}{}
		attrValuesByExportedName[exportedAttrName] = attrValue
	}
	if e.root && options.namedRoot {
		fmt.Fprintf(w, "%s\tXMLName xml.Name `xml:\"%s\"`\n", indentPrefix, e.name.Local)
	}
	for _, exportedAttrName := range slices.Sorted(maps.Keys(attrValuesByExportedName)) {
		attrValue := attrValuesByExportedName[exportedAttrName]
		fmt.Fprintf(w, "%s\t%s %s `xml:\"%s,attr\"`\n", indentPrefix, exportedAttrName, attrValue.goType(options), attrValue.name.Local)
	}

	if e.charDataValue.observations > 0 {
		fieldName := options.charDataFieldName
		if _, ok := fieldNames[fieldName]; ok {
			return fmt.Errorf("%s: duplicate field name", fieldName)
		}
		fieldNames[fieldName] = struct{}{}
		fmt.Fprintf(w, "%s\t%s string `xml:\",chardata\"`\n", indentPrefix, fieldName)
	}

	childElements := slices.Collect(maps.Values(e.childElements))
	if options.preserveOrder {
		slices.SortFunc(childElements, func(a, b *element) int {
			return e.childOrder[a.name] - e.childOrder[b.name]
		})
	} else {
		slices.SortFunc(childElements, func(a, b *element) int {
			aExportedName := exportedNameWithoutSuffix(a, options)
			bExportedName := exportedNameWithoutSuffix(b, options)
			switch {
			case aExportedName < bExportedName:
				return -1
			case aExportedName == bExportedName:
				// When exported names are equal, fall back to original XML element name
				aOriginalName := options.exportNameFunc(a.name)
				bOriginalName := options.exportNameFunc(b.name)
				switch {
				case aOriginalName < bOriginalName:
					return -1
				case aOriginalName == bOriginalName:
					return 0
				default:
					return 1
				}
			default:
				return 1
			}
		})
	}

	for _, childElement := range childElements {
		// Decide whether to use compact types for this specific element
		shouldCompact := options.compactTypes &&
			childElement.isContainer() &&
			!options.nonCompactableElements[childElement.name]

		// Create options for this element with appropriate compact setting
		elementOptions := *options
		if !shouldCompact {
			elementOptions.compactTypes = false
		}
		exportedChildName := exportedName(childElement, &elementOptions)

		if _, ok := fieldNames[exportedChildName]; ok {
			// Only report field name conflicts if we're not using named types for this element
			if _, hasNamedType := options.namedTypes[childElement.name]; !hasNamedType {
				return fmt.Errorf("%s: duplicate field name", exportedChildName)
			}
		}
		fieldNames[exportedChildName] = struct{}{}

		// Check for repeated/optional, considering compact types path.
		_, repeated := e.repeatedChildren[childElement.name]
		_, optional := e.optionalChildren[childElement.name]

		// For compact types, also check if any element along the compact path is repeated.
		if shouldCompact && !repeated {
			targetChild := firstNotContainerElement(childElement)
			if targetChild != childElement {
				repeated = isRepeatedInCompactPath(childElement, targetChild)
			}
		}

		fmt.Fprintf(w, "%s\t%s ", indentPrefix, exportedChildName)
		if repeated {
			fmt.Fprintf(w, "[]")
		} else if options.usePointersForOptionalFields {
			if optional {
				fmt.Fprintf(w, "*")
			}
		}

		currentChild := childElement
		if shouldCompact {
			currentChild = firstNotContainerElement(childElement)
		}
		if topLevelElement, ok := options.namedTypes[currentChild.name]; ok {
			fmt.Fprintf(w, "%s", options.exportTypeNameFunc(topLevelElement.name))
		} else if _, ok := options.simpleTypes[currentChild.name]; ok {
			fmt.Fprintf(w, "%s", currentChild.charDataValue.goType(options))
		} else {
			if err := currentChild.writeGoType(w, options, indentPrefix+"\t"); err != nil {
				return err
			}
		}
		fmt.Fprintf(w, " `xml:\"%s\"`\n", attrName(childElement, shouldCompact))
	}

	fmt.Fprintf(w, "%s}", indentPrefix)
	return nil
}

func (e *element) isContainer() bool {
	return len(e.childElements) == 1 && len(e.attrValues) == 0 && e.charDataValue.observations == 0
}

func firstNotContainerElement(el *element) *element {
	if el.isContainer() {
		for _, v := range el.childElements {
			if el == v {
				return el
			}
			return firstNotContainerElement(v)
		}
	}
	return el
}

// isRepeatedInCompactPath checks if any element along the compact path from start to target is repeated.
func isRepeatedInCompactPath(start, target *element) bool {
	current := start
	for current != target && current.isContainer() {
		for childName, childElement := range current.childElements {
			if _, repeated := current.repeatedChildren[childName]; repeated {
				return true
			}
			if childElement == target || isAncestorOf(childElement, target) {
				current = childElement
				break
			}
		}
	}
	return false
}

// isAncestorOf checks if ancestor is an ancestor of descendant in the element tree.
func isAncestorOf(ancestor, descendant *element) bool {
	if ancestor == descendant {
		return true
	}
	if ancestor.isContainer() {
		for _, child := range ancestor.childElements {
			if isAncestorOf(child, descendant) {
				return true
			}
		}
	}
	return false
}

func exportedName(el *element, options *generateOptions) string {
	return exportedNameWithoutSuffix(el, options) + options.elemNameSuffix
}

func exportedNameWithoutSuffix(el *element, options *generateOptions) string {
	if el.isContainer() && options.compactTypes {
		for _, v := range el.childElements {
			if el == v {
				return options.exportNameFunc(el.name)
			}
			return exportedNameWithoutSuffix(v, options)
		}
	}
	return options.exportNameFunc(el.name)
}

func attrName(el *element, compactTypes bool) string {
	if el.isContainer() && compactTypes {
		for _, v := range el.childElements {
			if el == v {
				s := el.name.Local + ">"
				s += strings.Repeat(s, el.nestedCount)
				return s[:len(s)-1]
			}
			return el.name.Local + ">" + attrName(v, true)
		}
	}
	return el.name.Local
}
