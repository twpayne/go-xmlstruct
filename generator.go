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
	packageName                  string
	nameFunc                     NameFunc
	timeLayout                   string
	usePointersForOptionalFields bool
	typeElements                 map[xml.Name]*element
}

// A GeneratorOption sets an option on a Generator.
type GeneratorOption func(*Generator)

// WithExportNameFunc sets the export name function for the generated Go source.
func WithExportNameFunc(exportNameFunc ExportNameFunc) GeneratorOption {
	return func(o *Generator) {
		o.exportNameFunc = exportNameFunc
	}
}

// WithFormatSource sets whether to format the generated Go source.
func WithFormatSource(formatSource bool) GeneratorOption {
	return func(o *Generator) {
		o.formatSource = formatSource
	}
}

// WithHeader sets the header of the generated Go source.
func WithHeader(header string) GeneratorOption {
	return func(o *Generator) {
		o.header = header
	}
}

// WithIntType sets the int type in the generated Go source.
func WithIntType(intType string) GeneratorOption {
	return func(o *Generator) {
		o.intType = intType
	}
}

// WithNameFunc sets the name function.
func WithNameFunc(nameFunc NameFunc) GeneratorOption {
	return func(o *Generator) {
		o.nameFunc = nameFunc
	}
}

// WithPackageName sets the package name of the generated Go source.
func WithPackageName(packageName string) GeneratorOption {
	return func(o *Generator) {
		o.packageName = packageName
	}
}

// WithTimeLayout sets the time layout used to identify times in the observed
// XML documents. Use an empty string to disable identifying times.
func WithTimeLayout(timeLayout string) GeneratorOption {
	return func(o *Generator) {
		o.timeLayout = timeLayout
	}
}

// WithUsePointersForOptionFields sets whether to use pointers for optional
// fields in the generated Go source.
func WithUsePointersForOptionalFields(usePointersForOptionalFields bool) GeneratorOption {
	return func(o *Generator) {
		o.usePointersForOptionalFields = usePointersForOptionalFields
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
		packageName:                  DefaultPackageName,
		timeLayout:                   DefaultTimeLayout,
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
	options := sourceOptions{
		exportNameFunc:               g.exportNameFunc,
		header:                       g.header,
		importPackageNames:           make(map[string]struct{}),
		intType:                      g.intType,
		usePointersForOptionalFields: g.usePointersForOptionalFields,
	}

	typesBuilder := &strings.Builder{}
	for _, typeName := range sortXMLNames(maps.Keys(g.typeElements)) {
		fmt.Fprintf(typesBuilder, "\ntype %s ", options.exportNameFunc(typeName))
		typeElement := g.typeElements[typeName]
		typeElement.writeGoType(typesBuilder, &options, "")
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
		nameFunc:   g.nameFunc,
		timeLayout: g.timeLayout,
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
				if err := typeElement.observeChildElement(decoder, startElement, &options); err != nil {
					return err
				}
			}
		}
	}
}
