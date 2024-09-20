# go-xmlstruct

[![PkgGoDev](https://pkg.go.dev/badge/github.com/twpayne/go-xmlstruct)](https://pkg.go.dev/github.com/twpayne/go-xmlstruct)

Generate Go structs from multiple XML documents.

## What does go-xmlstruct do and why should I use it?

go-xmlstruct generates Go structs from XML documents. Alternatively put,
go-xmlstruct infers XML schemas from one or more example XML documents. For
example, given [this XML
document](https://github.com/twpayne/go-xmlstruct/blob/master/internal/tests/play/testdata/all_well.xml),
go-xmlstruct generates [this Go source
code](https://github.com/twpayne/go-xmlstruct/blob/master/internal/tests/play/play.gen.go).

Compared to existing Go struct generators like
[zek](https://github.com/miku/zek),
[XMLGen](https://github.com/dutchcoders/XMLGen), and
[chidley](https://github.com/gnewton/chidley), go-xmlstruct:

* Takes multiple XML documents as input.
* Generates field types of `bool`, `int`, `string`, or `time.Time` as
  appropriate.
* Creates named types for all elements.
* Handles optional attributes and elements.
* Handles repeated attributes and elements.
* Ignores empty chardata.
* Provides a CLI for simple use.
* Usable as a Go package for advanced use, including configurable field naming.

go-xmlstruct is useful for quick-and-dirty unmarshalling of arbitrary XML
documents, especially when you have no schema or the schema is extremely complex
and you want something that "just works" with the documents you have.

## Install

Install the `goxmlstruct` CLI with:

```console
$ go install github.com/twpayne/go-xmlstruct/cmd/goxmlstruct@latest
```

## Example

Feed `goxmlstruct` the simple XML document:

```xml
<parent>
  <child flag="true">
    text
  </child>
</parent>
```

by running:

```console
$ echo '<parent><child flag="true">text</child></parent>' | goxmlstruct
```

This produces the output:

```go
// Code generated by goxmlstruct. DO NOT EDIT.

package main

type Parent struct {
        Child struct {
                Flag     bool   `xml:"flag,attr"`
                CharData string `xml:",chardata"`
        } `xml:"child"`
}
```

This demonstrates:

* A Go struct is generated from the structure of the input XML document.
* Attributes, child elements, and chardata are all considered.
* Field names are generated automatically.
* Field types are detected automatically.

For a full list of options to the `goxmlstruct` CLI run:

```console
$ goxmlstruct -help
```

You can run a more advanced example with:

```console
$ git clone https://github.com/twpayne/go-xmlstruct.git
$ cd go-xmlstruct
$ goxmlstruct internal/tests/gpx/testdata/*.gpx
```

This demonstrates generating a Go struct from multiple complex XML documents.

For an example of configurable field naming and named types by using
go-xmlstruct as a package, see
[`internal/tests/play/play_test.go`](https://github.com/twpayne/go-xmlstruct/blob/master/internal/tests/play/play_test.go).

For an example of a complex schema, see
[`internal/tests/aixm/aixm_test.go`](https://github.com/twpayne/go-xmlstruct/blob/master/internal/tests/aixm/aixm_test.go).

## How does go-xmlstruct work?

Similar to [go-jsonstruct](https://github.com/twpayne/go-jsonstruct), go-xmlstruct consists of two phases:

1. Firstly, go-xmlstruct explores all input XML documents to determine their
   structure. It gathers statistics on the types used for each attribute,
   chardata, and child element.
2. Secondly, go-xmlstruct generates a Go struct based on the observed structure
   using the gathered statistics to determine the type of each field.

## License

MIT