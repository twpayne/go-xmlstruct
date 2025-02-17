// Command goxmlstruct generates Go structs from multiple XML documents.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/spf13/pflag"

	"github.com/twpayne/go-xmlstruct"
)

var (
	attrNameSuffix               = pflag.String("attr-name-suffix", xmlstruct.DefaultAttrNameSuffix, "attribute name suffix")
	charDataFieldName            = pflag.String("char-data-field-name", xmlstruct.DefaultCharDataFieldName, "char data field name")
	compactTypes                 = pflag.Bool("compact-types", xmlstruct.DefaultCompactTypes, "create compact types")
	elemNameSuffix               = pflag.String("elem-name-suffix", xmlstruct.DefaultElemNameSuffix, "element name suffix")
	formatSource                 = pflag.Bool("format-source", xmlstruct.DefaultFormatSource, "format source")
	header                       = pflag.String("header", xmlstruct.DefaultHeader, "header")
	ignoreErrors                 = pflag.Bool("ignore-errors", false, "ignore errors")
	ignoreNamespaces             = pflag.Bool("ignore-namespaces", true, "ignore namespaces")
	imports                      = pflag.Bool("imports", xmlstruct.DefaultImports, "generate import statements")
	intType                      = pflag.String("int-type", xmlstruct.DefaultIntType, "int type")
	namedRoot                    = pflag.Bool("named-root", xmlstruct.DefaultNamedRoot, "create an XMLName field for the root element")
	namedTypes                   = pflag.Bool("named-types", xmlstruct.DefaultNamedTypes, "create named types for all elements")
	noEmptyElements              = pflag.Bool("no-empty-elements", !xmlstruct.DefaultEmptyElements, "use type string instead of struct{} for empty elements")
	noExport                     = pflag.Bool("no-export", false, "create unexported types")
	output                       = pflag.String("output", "", "output filename")
	packageName                  = pflag.String("package-name", "main", "package name")
	pattern                      = pflag.String("pattern", "", "filename pattern to observe")
	preserveOrder                = pflag.Bool("preserve-order", xmlstruct.DefaultPreserveOrder, "preserve order of types and fields")
	timeLayout                   = pflag.String("time-layout", "2006-01-02T15:04:05Z", "time layout")
	topLevelAttributes           = pflag.Bool("top-level-attributes", xmlstruct.DefaultTopLevelAttributes, "include top level attributes")
	typesOnly                    = pflag.Bool("types-only", false, "generate structs only, without header, package, or imports")
	usePointersForOptionalFields = pflag.Bool("use-pointers-for-optional-fields", xmlstruct.DefaultUsePointersForOptionalFields, "use pointers for optional fields")
	useRawToken                  = pflag.Bool("use-raw-token", xmlstruct.DefaultUseRawToken, "use encoding/xml.Decoder.RawToken")
)

func run() error {
	pflag.Parse()

	nameFunc := xmlstruct.IdentityNameFunc
	if *ignoreNamespaces {
		nameFunc = xmlstruct.IgnoreNamespaceNameFunc
	}

	if *typesOnly {
		*header = ""
		*imports = false
		*packageName = ""
	}

	options := []xmlstruct.GeneratorOption{
		xmlstruct.WithAttrNameSuffix(*attrNameSuffix),
		xmlstruct.WithCharDataFieldName(*charDataFieldName),
		xmlstruct.WithCompactTypes(*compactTypes),
		xmlstruct.WithElemNameSuffix(*elemNameSuffix),
		xmlstruct.WithEmptyElements(!*noEmptyElements),
		xmlstruct.WithFormatSource(*formatSource),
		xmlstruct.WithHeader(*header),
		xmlstruct.WithImports(*imports),
		xmlstruct.WithIntType(*intType),
		xmlstruct.WithNamedRoot(*namedRoot),
		xmlstruct.WithNamedTypes(*namedTypes),
		xmlstruct.WithNameFunc(nameFunc),
		xmlstruct.WithPackageName(*packageName),
		xmlstruct.WithPreserveOrder(*preserveOrder),
		xmlstruct.WithTimeLayout(*timeLayout),
		xmlstruct.WithTopLevelAttributes(*topLevelAttributes),
		xmlstruct.WithUsePointersForOptionalFields(*usePointersForOptionalFields),
		xmlstruct.WithUseRawToken(*useRawToken),
	}
	if *noExport {
		options = append(options, xmlstruct.WithExportTypeNameFunc(xmlstruct.DefaultUnexportNameFunc))
	}
	generator := xmlstruct.NewGenerator(options...)

	filenames := slices.Clone(pflag.Args())
	if *pattern != "" {
		matches, err := filepath.Glob(*pattern)
		if err != nil {
			return err
		}
		filenames = append(filenames, matches...)
	}

	if len(filenames) == 0 {
		if err := generator.ObserveReader(os.Stdin); err != nil {
			return err
		}
	} else {
		for _, filename := range filenames {
			if err := generator.ObserveFile(filename); err != nil {
				if *ignoreErrors {
					fmt.Fprintf(os.Stderr, "%s: %v\n", filename, err)
				} else {
					return fmt.Errorf("%s: %w", filename, err)
				}
			}
		}
	}

	source, err := generator.Generate()
	if err != nil {
		return err
	}

	if *output == "" {
		_, err := os.Stdout.Write(source)
		return err
	}
	return os.WriteFile(*output, source, 0o666)
}

func main() {
	if err := run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
