// Command xmlstruct generates Go structs from multiple XML documents.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/twpayne/go-xmlstruct"
)

var (
	charDataFieldName            = flag.String("char-data-field-name", xmlstruct.DefaultCharDataFieldName, "char data field name")
	formatSource                 = flag.Bool("format-source", xmlstruct.DefaultFormatSource, "format source")
	header                       = flag.String("header", xmlstruct.DefaultHeader, "header")
	ignoreNamespaces             = flag.Bool("ignore-namespaces", true, "ignore namespaces")
	intType                      = flag.String("int-type", xmlstruct.DefaultIntType, "int type")
	namedTypes                   = flag.Bool("named-types", xmlstruct.DefaultNamedTypes, "create named types for all elements")
	output                       = flag.String("output", "", "output filename")
	packageName                  = flag.String("package-name", "main", "package name")
	preserveOrder                = flag.Bool("preserve-order", xmlstruct.DefaultPreserveOrder, "preserve order of types and fields")
	timeLayout                   = flag.String("time-layout", "2006-01-02T15:04:05Z", "time layout")
	topLevelAttributes           = flag.Bool("top-level-attributes", xmlstruct.DefaultTopLevelAttributes, "include top level attributes")
	usePointersForOptionalFields = flag.Bool("use-pointers-for-optional-fields", xmlstruct.DefaultUsePointersForOptionalFields, "use pointers for optional fields")
)

func run() error {
	flag.Parse()

	nameFunc := xmlstruct.IdentityNameFunc
	if *ignoreNamespaces {
		nameFunc = xmlstruct.IgnoreNamespaceNameFunc
	}

	generator := xmlstruct.NewGenerator(
		xmlstruct.WithCharDataFieldName(*charDataFieldName),
		xmlstruct.WithFormatSource(*formatSource),
		xmlstruct.WithHeader(*header),
		xmlstruct.WithIntType(*intType),
		xmlstruct.WithNameFunc(nameFunc),
		xmlstruct.WithNamedTypes(*namedTypes),
		xmlstruct.WithPackageName(*packageName),
		xmlstruct.WithPreserveOrder(*preserveOrder),
		xmlstruct.WithTimeLayout(*timeLayout),
		xmlstruct.WithTopLevelAttributes(*topLevelAttributes),
		xmlstruct.WithUsePointersForOptionalFields(*usePointersForOptionalFields),
	)

	if flag.NArg() == 0 {
		if err := generator.ObserveReader(os.Stdin); err != nil {
			return err
		}
	} else {
		for _, arg := range flag.Args() {
			if err := generator.ObserveFile(arg); err != nil {
				return err
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
