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

const (
	DefaultPackageName = "main"
	DefaultTimeLayout  = "2006-01-02T15:04:05Z"
)

type Generator struct {
	observedElements             map[xml.Name]*observedElement
	exportNameFunc               ExportNameFunc
	formatSource                 bool
	intType                      string
	packageName                  string
	nameFunc                     NameFunc
	timeLayout                   string
	usePointersForOptionalFields bool
}

type GeneratorOption func(*Generator)

func WithExportNameFunc(exportNameFunc ExportNameFunc) GeneratorOption {
	return func(o *Generator) {
		o.exportNameFunc = exportNameFunc
	}
}

func WithFormatSource(formatSource bool) GeneratorOption {
	return func(o *Generator) {
		o.formatSource = formatSource
	}
}

func WithIntType(intType string) GeneratorOption {
	return func(o *Generator) {
		o.intType = intType
	}
}

func WithNameFunc(nameFunc NameFunc) GeneratorOption {
	return func(o *Generator) {
		o.nameFunc = nameFunc
	}
}

func WithPackageName(packageName string) GeneratorOption {
	return func(o *Generator) {
		o.packageName = packageName
	}
}

func WithTimeLayout(timeLayout string) GeneratorOption {
	return func(o *Generator) {
		o.timeLayout = timeLayout
	}
}

func WithUsePointersForOptionalFields(usePointersForOptionalFields bool) GeneratorOption {
	return func(o *Generator) {
		o.usePointersForOptionalFields = usePointersForOptionalFields
	}
}

func NewGenerator(options ...GeneratorOption) *Generator {
	generator := &Generator{
		observedElements:             make(map[xml.Name]*observedElement),
		exportNameFunc:               DefaultExportNameFunc,
		formatSource:                 true,
		intType:                      "int",
		nameFunc:                     IgnoreNamespaceNameFunc,
		usePointersForOptionalFields: true,
		packageName:                  DefaultPackageName,
		timeLayout:                   DefaultTimeLayout,
	}
	for _, option := range options {
		option(generator)
	}
	return generator
}

func (g *Generator) Generate() ([]byte, error) {
	options := sourceOptions{
		exportNameFunc:               g.exportNameFunc,
		importPackageNames:           make(map[string]struct{}),
		intType:                      g.intType,
		usePointersForOptionalFields: g.usePointersForOptionalFields,
	}

	typesBuilder := &strings.Builder{}
	for _, typeName := range sortXMLNames(maps.Keys(g.observedElements)) {
		fmt.Fprintf(typesBuilder, "\ntype %s ", options.exportNameFunc(typeName))
		observedElement := g.observedElements[typeName]
		observedElement.writeGoType(typesBuilder, &options, "")
		typesBuilder.WriteByte('\n')
	}

	sourceBuilder := &strings.Builder{}
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

func (g *Generator) ObserveFile(name string) error {
	file, err := os.Open(name)
	if err != nil {
		return err
	}
	defer file.Close()
	return g.ObserveReader(file)
}

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
				observedElement, ok := g.observedElements[name]
				if !ok {
					observedElement = newObservedElement(name)
					g.observedElements[name] = observedElement
				}
				if err := observedElement.observeChildElement(decoder, startElement, &options); err != nil {
					return err
				}
			}
		}
	}
}
