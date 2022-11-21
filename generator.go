package xmlstruct

import (
	"encoding/xml"
	"errors"
	"fmt"
	"go/format"
	"io"
	"os"
	"sort"
	"strings"

	"golang.org/x/exp/maps"
	"golang.org/x/net/html/charset"
)

// A Generator observes XML documents and generates Go structs into which the
// XML documents can be unmarshalled.
type Generator struct {
	exportNameFunc               ExportNameFunc
	formatSource                 bool
	header                       string
	intType                      string
	nameFunc                     NameFunc
	namedTypes                   bool
	packageName                  string
	timeLayout                   string
	topLevelAttributes           bool
	usePointersForOptionalFields bool
	typeElements                 map[xml.Name]*element
}

// A GeneratorOption sets an option on a Generator.
type GeneratorOption func(*Generator)

// WithExportNameFunc sets the export name function for the generated Go source.
func WithExportNameFunc(exportNameFunc ExportNameFunc) GeneratorOption {
	return func(g *Generator) {
		g.exportNameFunc = exportNameFunc
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

// NewGenerator returns a new Generator with the given options.
func NewGenerator(options ...GeneratorOption) *Generator {
	generator := &Generator{
		exportNameFunc:               DefaultExportNameFunc,
		formatSource:                 DefaultFormatSource,
		header:                       DefaultHeader,
		intType:                      DefaultIntType,
		nameFunc:                     DefaultNameFunc,
		namedTypes:                   DefaultNamedTypes,
		packageName:                  DefaultPackageName,
		timeLayout:                   DefaultTimeLayout,
		topLevelAttributes:           DefaultTopLevelAttributes,
		usePointersForOptionalFields: DefaultUsePointersForOptionalFields,
		typeElements:                 make(map[xml.Name]*element),
	}
	for _, option := range options {
		option(generator)
	}
	return generator
}

// Generate returns the generated Go source for all the XML documents observed
// so far.
func (g *Generator) Generate() ([]byte, error) {
	options := generateOptions{
		exportNameFunc:               g.exportNameFunc,
		header:                       g.header,
		importPackageNames:           make(map[string]struct{}),
		intType:                      g.intType,
		usePointersForOptionalFields: g.usePointersForOptionalFields,
	}

	typeElements := maps.Clone(g.typeElements)
	if g.namedTypes {
		options.namedTypes = typeElements
		options.simpleTypes = make(map[xml.Name]struct{})
		for name, element := range options.namedTypes {
			if len(element.attrValues) != 0 || len(element.childElements) != 0 {
				continue
			}
			options.simpleTypes[name] = struct{}{}
			delete(options.namedTypes, name)
		}
	}

	typesBuilder := &strings.Builder{}
	typeElementsByExportedName := make(map[string]*element, len(typeElements))
	for typeName, typeElement := range typeElements {
		exportedName := options.exportNameFunc(typeName)
		if _, ok := typeElementsByExportedName[exportedName]; ok {
			return nil, fmt.Errorf("%s: duplicate type name", exportedName)
		}
		typeElementsByExportedName[exportedName] = typeElement
	}
	for _, exportedName := range sortedKeys(typeElementsByExportedName) {
		fmt.Fprintf(typesBuilder, "\ntype %s ", exportedName)
		typeElement := typeElementsByExportedName[exportedName]
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
		nameFunc:           g.nameFunc,
		timeLayout:         g.timeLayout,
		topLevelAttributes: g.topLevelAttributes,
	}
	if g.namedTypes {
		options.topLevelElements = g.typeElements
	}
	decoder := xml.NewDecoder(r)
	decoder.CharsetReader = charset.NewReaderLabel
	for {
		switch token, err := decoder.Token(); {
		case errors.Is(err, io.EOF):
			return nil
		case err != nil:
			return err
		default:
			if startElement, ok := token.(xml.StartElement); ok {
				name := g.nameFunc(startElement.Name)
				typeElement, ok := g.typeElements[name]
				if !ok {
					typeElement = newElement(name)
					g.typeElements[name] = typeElement
				}
				if err := typeElement.observeChildElement(decoder, startElement, 0, &options); err != nil {
					return err
				}
			}
		}
	}
}
