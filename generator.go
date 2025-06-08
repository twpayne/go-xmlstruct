package xmlstruct

import (
	"encoding/xml"
	"errors"
	"fmt"
	"go/format"
	"io"
	"io/fs"
	"os"
	"slices"
	"sort"
	"strings"

	"golang.org/x/net/html/charset"
)

var (
	SkipDir = fs.SkipDir //nolint:errname
	//lint:ignore ST1012 SkipFile is not an error
	SkipFile = errors.New("skip file") //nolint:errname,revive
)

// A ModifyDecoderFunc makes arbitrary changes to an encoding/xml.Decoder before
// it is used.
type ModifyDecoderFunc func(*xml.Decoder)

// A Generator observes XML documents and generates Go structs into which the
// XML documents can be unmarshalled.
type Generator struct {
	attrNameSuffix               string
	charDataFieldName            string
	elemNameSuffix               string
	exportNameFunc               ExportNameFunc
	exportTypeNameFunc           ExportNameFunc
	exportRenames                map[string]string
	formatSource                 bool
	header                       string
	imports                      bool
	intType                      string
	modifyDecoderFunc            ModifyDecoderFunc
	nameFunc                     NameFunc
	namedRoot                    bool
	namedTypes                   bool
	compactTypes                 bool
	order                        int
	packageName                  string
	preserveOrder                bool
	timeLayout                   string
	topLevelAttributes           bool
	typeOrder                    map[xml.Name]int
	usePointersForOptionalFields bool
	useRawToken                  bool
	typeElements                 map[xml.Name]*element
	emptyElements                bool
}

// A GeneratorOption sets an option on a Generator.
type GeneratorOption func(*Generator)

// WithAttrNameSuffix sets the attribute suffix.
func WithAttrNameSuffix(attrSuffix string) GeneratorOption {
	return func(g *Generator) {
		g.attrNameSuffix = attrSuffix
	}
}

// WithCharDataFieldName sets the char data field name.
func WithCharDataFieldName(charDataFieldName string) GeneratorOption {
	return func(g *Generator) {
		g.charDataFieldName = charDataFieldName
	}
}

// WithElemNameSuffix sets the element name suffix.
func WithElemNameSuffix(elemNameSuffix string) GeneratorOption {
	return func(g *Generator) {
		g.elemNameSuffix = elemNameSuffix
	}
}

// WithEmptyElements sets whether to use type struct{} or string
// for empty xml elements.
func WithEmptyElements(emptyElements bool) GeneratorOption {
	return func(g *Generator) {
		g.emptyElements = emptyElements
	}
}

// WithExportNameFunc sets the export name function for the generated Go source.
// It overrides WithExportRenames.
func WithExportNameFunc(exportNameFunc ExportNameFunc) GeneratorOption {
	return func(g *Generator) {
		g.exportNameFunc = exportNameFunc
	}
}

// WithExportTypeNameFunc sets the export name function for the generated Go source Types.
// This is useful when unexported types are desired.
func WithExportTypeNameFunc(exportTypeNameFunc ExportNameFunc) GeneratorOption {
	return func(g *Generator) {
		g.exportTypeNameFunc = exportTypeNameFunc
	}
}

// WithExportRenames sets the export renames. It is overridden by
// WithExportRenameFunc.
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

// WithImports sets whether to include an import statement in the generated code.
func WithImports(withImports bool) GeneratorOption {
	return func(g *Generator) {
		g.imports = withImports
	}
}

// WithIntType sets the int type in the generated Go source.
func WithIntType(intType string) GeneratorOption {
	return func(g *Generator) {
		g.intType = intType
	}
}

// WithModifyDecoderFunc sets the function that will modify the
// encoding/xml.Decoder used.
func WithModifyDecoderFunc(modifyDecoderFunc ModifyDecoderFunc) GeneratorOption {
	return func(g *Generator) {
		g.modifyDecoderFunc = modifyDecoderFunc
	}
}

// WithNameFunc sets the name function.
func WithNameFunc(nameFunc NameFunc) GeneratorOption {
	return func(g *Generator) {
		g.nameFunc = nameFunc
	}
}

// WithNamedRoot sets whether to generate an XMLName field for the root element.
func WithNamedRoot(namedRoot bool) GeneratorOption {
	return func(o *Generator) {
		o.namedRoot = namedRoot
	}
}

// WithNamedTypes sets whether all to generate named types for all elements.
func WithNamedTypes(namedTypes bool) GeneratorOption {
	return func(o *Generator) {
		o.namedTypes = namedTypes
	}
}

// WithCompactTypes sets whether to generate compact types.
func WithCompactTypes(compactTypes bool) GeneratorOption {
	return func(o *Generator) {
		o.compactTypes = compactTypes
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

// NewGenerator returns a new Generator with the given options.
func NewGenerator(options ...GeneratorOption) *Generator {
	g := &Generator{
		attrNameSuffix:               DefaultAttrNameSuffix,
		charDataFieldName:            DefaultCharDataFieldName,
		elemNameSuffix:               DefaultElemNameSuffix,
		formatSource:                 DefaultFormatSource,
		header:                       DefaultHeader,
		imports:                      DefaultImports,
		intType:                      DefaultIntType,
		nameFunc:                     DefaultNameFunc,
		namedRoot:                    DefaultNamedRoot,
		namedTypes:                   DefaultNamedTypes,
		compactTypes:                 DefaultCompactTypes,
		packageName:                  DefaultPackageName,
		preserveOrder:                DefaultPreserveOrder,
		timeLayout:                   DefaultTimeLayout,
		topLevelAttributes:           DefaultTopLevelAttributes,
		typeOrder:                    make(map[xml.Name]int),
		usePointersForOptionalFields: DefaultUsePointersForOptionalFields,
		useRawToken:                  DefaultUseRawToken,
		typeElements:                 make(map[xml.Name]*element),
		emptyElements:                DefaultEmptyElements,
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
	if g.exportTypeNameFunc == nil {
		g.exportTypeNameFunc = g.exportNameFunc
	}
	return g
}

// Generate returns the generated Go source for all the XML documents observed
// so far.
func (g *Generator) Generate() ([]byte, error) {
	options := generateOptions{
		attrNameSuffix:               g.attrNameSuffix,
		charDataFieldName:            g.charDataFieldName,
		elemNameSuffix:               g.elemNameSuffix,
		exportNameFunc:               g.exportNameFunc,
		exportTypeNameFunc:           g.exportTypeNameFunc,
		header:                       g.header,
		importPackageNames:           make(map[string]struct{}),
		intType:                      g.intType,
		namedRoot:                    g.namedRoot,
		compactTypes:                 g.compactTypes,
		preserveOrder:                g.preserveOrder,
		usePointersForOptionalFields: g.usePointersForOptionalFields,
		emptyElements:                g.emptyElements,
	}

	if options.namedRoot {
		options.importPackageNames["encoding/xml"] = struct{}{}
	}

	var typeElements []*element
	if g.namedTypes {
		options.namedTypes = make(map[xml.Name]*element)
		for k, v := range g.typeElements {
			if !options.compactTypes || !v.isContainer() || v.root {
				options.namedTypes[k] = v
			}
		}
		options.simpleTypes = make(map[xml.Name]struct{})
		for name, element := range options.namedTypes {
			if len(element.attrValues) != 0 || len(element.childElements) != 0 || element.root {
				continue
			}
			options.simpleTypes[name] = struct{}{}
			delete(options.namedTypes, name)
		}
		typeElements = mapValues(options.namedTypes)
	} else {
		typeElements = mapValues(g.typeElements)
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
		typeName := options.exportTypeNameFunc(typeElement.name)
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
	packageName := g.packageName
	if packageName == "" {
		packageName = "main"
	}
	packageDeclaration := "package " + packageName + "\n"
	sourceBuilder.WriteString(packageDeclaration)
	if g.imports {
		switch len(options.importPackageNames) {
		case 0:
			// Do nothing.
		case 1:
			for importPackageName := range options.importPackageNames {
				fmt.Fprintf(sourceBuilder, "import %q\n", importPackageName)
			}
		default:
			fmt.Fprintf(sourceBuilder, "import (\n")
			importPackageNames := mapKeys(options.importPackageNames)
			sort.Strings(importPackageNames)
			for _, importPackageName := range importPackageNames {
				fmt.Fprintf(sourceBuilder, "\t%q\n", importPackageName)
			}
			fmt.Fprintf(sourceBuilder, ")\n")
		}
	}
	sourceBuilder.WriteString(typesBuilder.String())

	source := []byte(sourceBuilder.String())
	if g.formatSource {
		if formattedSource, err := format.Source(source); err == nil {
			source = formattedSource
		}
	}
	if g.packageName == "" {
		indexOfPackageDeclaration := 0
		if g.header != "" {
			indexOfPackageDeclaration = len(g.header) + 2
		}
		sourceWithoutPackageDeclaration := make([]byte, 0, len(source))
		sourceWithoutPackageDeclaration = append(sourceWithoutPackageDeclaration, source[:indexOfPackageDeclaration]...)
		indexOfTypeDecleration := indexOfPackageDeclaration + len(packageDeclaration)
		// remove \n prefix
		if len(source) > indexOfTypeDecleration {
			indexOfTypeDecleration++
		}
		sourceWithoutPackageDeclaration = append(sourceWithoutPackageDeclaration, source[indexOfTypeDecleration:]...)
		source = sourceWithoutPackageDeclaration
	}

	return source, nil
}

// ObserveFS observes all files in fs.
//
// observeFunc is called before each entry. If observeFunc returns [fs.SkipDir]
// or [SkipFile] then the entry is skipped. If observeFunc returns any other
// non-nil error then ObserveFS terminates with the returned error. If observe
// func returns nil and the entry is a regular file or symlink then it is
// observed, otherwise the entry is ignored.
func (g *Generator) ObserveFS(fsys fs.FS, root string, observeFunc func(string, fs.DirEntry, error) error) error {
	return fs.WalkDir(fsys, root, func(path string, dirEntry fs.DirEntry, err error) error {
		switch err := observeFunc(path, dirEntry, err); {
		case errors.Is(err, fs.SkipDir):
			return fs.SkipDir
		case errors.Is(err, SkipFile):
			return nil
		case err != nil:
			return err
		case dirEntry.IsDir():
			return nil
		case dirEntry.Type() == 0 || dirEntry.Type() == fs.ModeSymlink:
			file, err := fsys.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()
			if err := g.ObserveReader(file); err != nil {
				return fmt.Errorf("%s: %w", path, err)
			}
			return nil
		default:
			return nil
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
	if g.modifyDecoderFunc != nil {
		g.modifyDecoderFunc(decoder)
	}
	var foundRootElement bool
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
				var root bool
				if !foundRootElement {
					foundRootElement = true
					root = true
				}
				name := g.nameFunc(startElement.Name)
				if name == (xml.Name{}) {
					continue FOR
				}
				typeElement, ok := g.typeElements[name]
				if !ok {
					typeElement = newElement(name)
					typeElement.root = root
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
