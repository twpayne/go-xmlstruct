// Command xmlstruct generates Go structs from multiple XML documents.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/twpayne/go-xmlstruct"
)

var (
	formatSource                 = flag.Bool("format-source", true, "format source")
	ignoreNamespaces             = flag.Bool("ignore-namespaces", true, "ignore namespaces")
	usePointersForOptionalFields = flag.Bool("use-pointers-for-optional-fields", true, "use pointers for optional fields")
	output                       = flag.String("output", "", "output filename")
	packageName                  = flag.String("package-name", "main", "package name")
	timeLayout                   = flag.String("time-layout", "2006-01-02T15:04:05Z", "time layout")
)

func run() error {
	flag.Parse()

	nameFunc := xmlstruct.IdentityNameFunc
	if *ignoreNamespaces {
		nameFunc = xmlstruct.IgnoreNamespaceNameFunc
	}

	generator := xmlstruct.NewGenerator(
		xmlstruct.WithFormatSource(*formatSource),
		xmlstruct.WithUsePointersForOptionalFields(*usePointersForOptionalFields),
		xmlstruct.WithNameFunc(nameFunc),
		xmlstruct.WithPackageName(*packageName),
		xmlstruct.WithTimeLayout(*timeLayout),
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
