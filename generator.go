package xmlstruct

import (
	"encoding/xml"
	"errors"
	"fmt"
	"go/format"
	"io"
	"io/fs"
	"os"
	"sort"
	"strings"

	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
	"golang.org/x/net/html/charset"
)

var (
	SkipDir = fs.SkipDir
	//lint:ignore ST1012 SkipFile is not an error
	SkipFile = errors.New("skip file") //nolint:errname

	unexpectedElement, unexpectedElements *element
)

// A Generator observes XML documents and generates Go structs into which the
// XML documents can be unmarshalled.
type Generator struct {
	charDataFieldName            string
	exportNameFunc               ExportNameFunc
	exportRenames                map[string]string
	formatSource                 bool
	header                       string
	intType                      string
	nameFunc                     NameFunc
	namedTypes                   bool
	order                        int
	packageName                  string
	preserveOrder                bool
	timeLayout                   string
	topLevelAttributes           bool
	typeOrder                    map[xml.Name]int
	usePointersForOptionalFields bool
	useRawToken                  bool
	supportUnexpectedElements    bool
	unexpectedElementTypeName    string
	typeElements                 map[xml.Name]*element
}

// A GeneratorOption sets an option on a Generator.
type GeneratorOption func(*Generator)

// WithCharDataFieldName sets the char data field name.
func WithCharDataFieldName(charDataFieldName string) GeneratorOption {
	return func(g *Generator) {
		g.charDataFieldName = charDataFieldName
	}
}

// WithExportNameFunc sets the export name function for the generated Go source.
func WithExportNameFunc(exportNameFunc ExportNameFunc) GeneratorOption {
	return func(g *Generator) {
		g.exportNameFunc = exportNameFunc
	}
}

// WithExportRenames sets the export renames.
func WithExportRenames(exportRenames map[string]string) GeneratorOption {
	return func(g *Generator) {
		g.exportRenames = exportRenames
	}
}

// WithFormatSource sets whether to format the generated Go source.
func WithFormatSource(formatSource bool) GeneratorOption {
	return func(g *Generator) {
		g.formatSource = formatSource
	}
}

// WithHeader sets the header of the generated Go source.
func WithHeader(header string) GeneratorOption {
	return func(g *Generator) {
		g.header = header
	}
}

// WithIntType sets the int type in the generated Go source.
func WithIntType(intType string) GeneratorOption {
	return func(g *Generator) {
		g.intType = intType
	}
}

// WithNameFunc sets the name function.
func WithNameFunc(nameFunc NameFunc) GeneratorOption {
	return func(g *Generator) {
		g.nameFunc = nameFunc
	}
}

// WithNamedTypes sets whether all to generate named types for all elements.
func WithNamedTypes(namedTypes bool) GeneratorOption {
	return func(o *Generator) {
		o.namedTypes = namedTypes
	}
}

// WithPackageName sets the package name of the generated Go source.
func WithPackageName(packageName string) GeneratorOption {
	return func(g *Generator) {
		g.packageName = packageName
	}
}

// WithPreserveOrder sets whether to preserve the order of types and fields.
func WithPreserveOrder(preserveOrder bool) GeneratorOption {
	return func(g *Generator) {
		g.preserveOrder = preserveOrder
	}
}

// WithTimeLayout sets the time layout used to identify times in the observed
// XML documents. Use an empty string to disable identifying times.
func WithTimeLayout(timeLayout string) GeneratorOption {
	return func(g *Generator) {
		g.timeLayout = timeLayout
	}
}

// WithTopLevelAttributes sets whether to include top level attributes.
func WithTopLevelAttributes(topLevelAttributes bool) GeneratorOption {
	return func(g *Generator) {
		g.topLevelAttributes = topLevelAttributes
	}
}

// WithUsePointersForOptionFields sets whether to use pointers for optional
// fields in the generated Go source.
func WithUsePointersForOptionalFields(usePointersForOptionalFields bool) GeneratorOption {
	return func(g *Generator) {
		g.usePointersForOptionalFields = usePointersForOptionalFields
	}
}

// WithUseRawToken sets whether to use encoding/xml.Decoder.Token or
// encoding/xml.Decoder.RawToken.
func WithUseRawToken(useRawToken bool) GeneratorOption {
	return func(g *Generator) {
		g.useRawToken = useRawToken
	}
}

// WithSupportUnexpectedElements sets whether to support unexpected elements
// on structs through use of a field of the catch-all type Node and xml:any
// struct tag on each generated struct.
func WithSupportUnexpectedElements(supportUnexpectedElements bool) GeneratorOption {
	return func(g *Generator) {
		g.supportUnexpectedElements = supportUnexpectedElements
	}
}

// WithUnexpectedElementTypeName specifies the name of the named type to contain
// any unexpected elements encountered during parsing.
func WithUnexpectedElementTypeName(unexpectedElementTypeName string) GeneratorOption {
	return func(g *Generator) {
		g.unexpectedElementTypeName = unexpectedElementTypeName
	}
}

// NewGenerator returns a new Generator with the given options.
func NewGenerator(options ...GeneratorOption) *Generator {
	g := &Generator{
		charDataFieldName:            DefaultCharDataFieldName,
		formatSource:                 DefaultFormatSource,
		header:                       DefaultHeader,
		intType:                      DefaultIntType,
		nameFunc:                     DefaultNameFunc,
		namedTypes:                   DefaultNamedTypes,
		packageName:                  DefaultPackageName,
		preserveOrder:                DefaultPreserveOrder,
		timeLayout:                   DefaultTimeLayout,
		topLevelAttributes:           DefaultTopLevelAttributes,
		typeOrder:                    make(map[xml.Name]int),
		usePointersForOptionalFields: DefaultUsePointersForOptionalFields,
		useRawToken:                  DefaultUseRawToken,
		supportUnexpectedElements:    DefaultSupportUnexpectedElements,
		unexpectedElementTypeName:    DefaultUnexpectedElementTypeName,
		typeElements:                 make(map[xml.Name]*element),
	}
	g.exportNameFunc = func(name xml.Name) string {
		if exportRename, ok := g.exportRenames[name.Local]; ok {
			return exportRename
		}
		return DefaultExportNameFunc(name)
	}
	for _, option := range options {
		option(g)
	}
	return g
}

// Generate returns the generated Go source for all the XML documents observed
// so far.
func (g *Generator) Generate() ([]byte, error) {
	options := generateOptions{
		charDataFieldName:            g.charDataFieldName,
		exportNameFunc:               g.exportNameFunc,
		header:                       g.header,
		importPackageNames:           make(map[string]struct{}),
		intType:                      g.intType,
		preserveOrder:                g.preserveOrder,
		usePointersForOptionalFields: g.usePointersForOptionalFields,
		supportUnexpectedElements:    g.supportUnexpectedElements,
		unexpectedElementTypeName:    g.unexpectedElementTypeName,
	}

	var typeElements []*element

	if g.supportUnexpectedElements {
		initializeUnexpectedElements(g.unexpectedElementTypeName)
		options.importPackageNames["encoding/xml"] = struct{}{}
		options.namedTypes = make(map[xml.Name]*element)
		options.namedTypes[xml.Name{Local: g.unexpectedElementTypeName}] = unexpectedElement
		g.typeElements[xml.Name{Local: g.unexpectedElementTypeName}] = unexpectedElement
		appendUnexpectedElements(&g.typeElements, g.namedTypes, g.unexpectedElementTypeName)
	}

	if g.namedTypes {
		options.namedTypes = maps.Clone(g.typeElements)
		options.simpleTypes = make(map[xml.Name]struct{})
		for name, element := range options.namedTypes {
			if len(element.attrValues) != 0 || len(element.childElements) != 0 {
				continue
			}
			options.simpleTypes[name] = struct{}{}
			delete(options.namedTypes, name)
		}
		typeElements = maps.Values(options.namedTypes)
	} else {
		typeElements = maps.Values(g.typeElements)
	}

	if options.preserveOrder {
		slices.SortFunc(typeElements, func(a, b *element) int {
			return g.typeOrder[a.name] - g.typeOrder[b.name]
		})
	} else {
		slices.SortFunc(typeElements, func(a, b *element) int {
			aExportedName := options.exportNameFunc(a.name)
			bExportedName := options.exportNameFunc(b.name)
			switch {
			// Force unexpected element struct to the top of the source
			case aExportedName == g.unexpectedElementTypeName && bExportedName != g.unexpectedElementTypeName:
				return -1
			case aExportedName == g.unexpectedElementTypeName && bExportedName == g.unexpectedElementTypeName:
				return 0
			case aExportedName != g.unexpectedElementTypeName && bExportedName == g.unexpectedElementTypeName:
				return 1
			case aExportedName < bExportedName:
				return -1
			case aExportedName == bExportedName:
				return 0
			default:
				return 1
			}
		})
	}

	typesBuilder := &strings.Builder{}
	typeNames := make(map[string]struct{})
	for _, typeElement := range typeElements {
		typeName := options.exportNameFunc(typeElement.name)
		if _, ok := typeNames[typeName]; ok {
			return nil, fmt.Errorf("%s: duplicate type name", typeName)
		}
		typeNames[typeName] = struct{}{}
		fmt.Fprintf(typesBuilder, "\ntype %s ", typeName)
		if err := typeElement.writeGoType(typesBuilder, &options, ""); err != nil {
			return nil, err
		}
		typesBuilder.WriteByte('\n')
	}

	sourceBuilder := &strings.Builder{}
	if options.header != "" {
		fmt.Fprintf(sourceBuilder, "%s\n\n", options.header)
	}
	fmt.Fprintf(sourceBuilder, "package %s\n\n", g.packageName)
	switch len(options.importPackageNames) {
	case 0:
		// Do nothing.
	case 1:
		for importPackageName := range options.importPackageNames {
			fmt.Fprintf(sourceBuilder, "import %q\n", importPackageName)
		}
	default:
		fmt.Fprintf(sourceBuilder, "import (\n")
		importPackageNames := maps.Keys(options.importPackageNames)
		sort.Strings(importPackageNames)
		for _, importPackageName := range importPackageNames {
			fmt.Fprintf(sourceBuilder, "\t%q\n", importPackageName)
		}
		fmt.Fprintf(sourceBuilder, ")\n")
	}
	sourceBuilder.WriteString(typesBuilder.String())

	source := []byte(sourceBuilder.String())
	if !g.formatSource {
		return source, nil
	}
	return format.Source(source)
}

// ObserveFS observes all XML documents in fs.
func (g *Generator) ObserveFS(fsys fs.FS, root string, observeFunc func(string, fs.DirEntry, error) error) error {
	return fs.WalkDir(fsys, root, func(path string, dirEntry fs.DirEntry, err error) error {
		switch err := observeFunc(path, dirEntry, err); {
		case errors.Is(err, fs.SkipDir):
			return fs.SkipDir
		case errors.Is(err, SkipFile):
			return nil
		case dirEntry.IsDir():
			return nil
		default:
			file, err := fsys.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()
			return g.ObserveReader(file)
		}
	})
}

// ObserveFile observes an XML document in the given file.
func (g *Generator) ObserveFile(name string) error {
	file, err := os.Open(name)
	if err != nil {
		return err
	}
	defer file.Close()
	return g.ObserveReader(file)
}

// ObserveReader observes an XML document from r.
func (g *Generator) ObserveReader(r io.Reader) error {
	options := observeOptions{
		getOrder: func() int {
			g.order++
			return g.order
		},
		nameFunc:           g.nameFunc,
		timeLayout:         g.timeLayout,
		topLevelAttributes: g.topLevelAttributes,
		typeOrder:          g.typeOrder,
		useRawToken:        g.useRawToken,
	}
	if g.namedTypes {
		options.topLevelElements = g.typeElements
	}

	decoder := xml.NewDecoder(r)
	decoder.CharsetReader = charset.NewReaderLabel
FOR:
	for {
		var token xml.Token
		var err error
		if g.useRawToken {
			token, err = decoder.RawToken()
		} else {
			token, err = decoder.Token()
		}
		switch {
		case errors.Is(err, io.EOF):
			return nil
		case err != nil:
			return err
		default:
			if startElement, ok := token.(xml.StartElement); ok {
				name := g.nameFunc(startElement.Name)
				if name == (xml.Name{}) {
					continue FOR
				}
				typeElement, ok := g.typeElements[name]
				if !ok {
					typeElement = newElement(name)
					g.typeElements[name] = typeElement
				}
				if _, ok := g.typeOrder[name]; !ok {
					g.typeOrder[name] = options.getOrder()
				}
				if err := typeElement.observeChildElement(decoder, startElement, 0, &options); err != nil {
					return err
				}
			}
		}
	}
}

func appendUnexpectedElements(elementsMap *map[xml.Name]*element, createNamedTypes bool, unexpectedElementTypeName string) {
	unexpectedElementsTypeName := fmt.Sprintf("%ss", unexpectedElementTypeName)
	for _, element := range *elementsMap {
		if element.name.Local == unexpectedElementTypeName {
			continue
		}
		if !createNamedTypes {
			appendUnexpectedElements(&element.childElements, createNamedTypes, unexpectedElementTypeName)
		}
		element.childElements[xml.Name{Local: unexpectedElementsTypeName}] = unexpectedElements
		element.repeatedChildren[xml.Name{Local: unexpectedElementsTypeName}] = struct{}{}
	}
}

func initializeUnexpectedElements(unexpectedElementTypeName string) {
	unexpectedElementTypeNameWithPointerSymbol := fmt.Sprintf("*%s", unexpectedElementTypeName)
	UnexpectedElementsTypeName := fmt.Sprintf("%ss", unexpectedElementTypeName)
	unexpectedElement = &element{
		name: xml.Name{Local: unexpectedElementTypeName},
		childElements: map[xml.Name]*element{
			{Local: "XMLName"}: {
				name: xml.Name{Local: "XMLName"},
				charDataValue: value{
					unexpectedElementTypeName: "xml.Name",
				},
			},
			{Local: "Attrs"}: {
				name: xml.Name{Local: "Attrs"},
				charDataValue: value{
					unexpectedElementTypeName: "xml.Attr",
				},
			},
			{Local: "Content"}: {
				name: xml.Name{Local: "Content"},
				charDataValue: value{
					unexpectedElementTypeName: "[]byte",
				},
			},
			{Local: "Nodes"}: {
				name: xml.Name{Local: "Nodes"},
				charDataValue: value{
					unexpectedElementTypeName: unexpectedElementTypeNameWithPointerSymbol,
				},
			},
		},
		repeatedChildren: map[xml.Name]struct{}{
			{Local: "Attrs"}: {},
			{Local: "Nodes"}: {},
		},
	}
	unexpectedElements = &element{
		name: xml.Name{Local: UnexpectedElementsTypeName},
		charDataValue: value{
			unexpectedElementTypeName: unexpectedElementTypeNameWithPointerSymbol,
		},
	}
}
