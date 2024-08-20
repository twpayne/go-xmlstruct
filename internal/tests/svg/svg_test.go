package svg

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/alecthomas/assert/v2"

	"github.com/twpayne/go-xmlstruct"
)

func TestSVG(t *testing.T) {
	t.Parallel()

	generator := xmlstruct.NewGenerator(
		xmlstruct.WithElemNameSuffix("Elem"),
		xmlstruct.WithExportRenames(map[string]string{
			"rx":                "RX",
			"ry":                "RY",
			"dx":                "DX",
			"dy":                "DY",
			"tref":              "TRef",
			"bbox":              "BBox",
			"bidi":              "BiDi",
			"cx":                "CX",
			"cy":                "CY",
			"fx":                "FX",
			"fy":                "FY",
			"hkern":             "HKern",
			"href":              "HRef",
			"li":                "LI",
			"lu":                "LU",
			"mpath":             "MPath",
			"ol":                "OL",
			"onactivate":        "OnActivate",
			"onbegin":           "OnBegin",
			"onclick":           "OnClick",
			"onend":             "OnEnd",
			"onfocusin":         "OnFocusIn",
			"onfocusout":        "OnFocusOut",
			"onmousedown":       "OnMouseDown",
			"onmousemove":       "OnMouseMove",
			"onmouseout":        "OnMouseOut",
			"onmouseover":       "OnMouseOver",
			"onmouseup":         "OnMouseUp",
			"operatorScript":    "OperatorScriptLowerElem",
			"polyline":          "PolyLine",
			"rdfs":              "RDFS",
			"stemh":             "StemH",
			"stemv":             "StemV",
			"stroke-dasharray":  "StrokeDashArray",
			"stroke-dashoffset": "StrokeDashOffset",
			"stroke-linecap":    "StrokeLineCap",
			"stroke-linejoin":   "StrokeLineJoin",
			"stroke-miterlimit": "StrokeMiterLimit",
			"svg":               "SVG",
			"testname":          "TestName",
			"tspan":             "TSpan",
			"ul":                "UL",
			"unicode-bidi":      "UnicodeBiDi",
			"xlink":             "XLink",
			"xmlns":             "XMLNS",
		}),
		xmlstruct.WithNamedTypes(true),
		xmlstruct.WithPackageName("svg"),
	)

	file, err := os.Open("testdata/W3C_SVG_11_TestSuite.tar.gz")
	assert.NoError(t, err)
	defer file.Close()

	gzipReader, err := gzip.NewReader(file)
	assert.NoError(t, err)
	defer gzipReader.Close()

	tarReader := tar.NewReader(gzipReader)
	for {
		header, err := tarReader.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		assert.NoError(t, err)
		if header.Typeflag != tar.TypeReg {
			continue
		}
		if !strings.HasSuffix(header.Name, ".svg") {
			continue
		}
		if err := generator.ObserveReader(tarReader); err != nil {
			t.Logf("%s: %v", header.Name, err)
		}
	}

	actualSource, err := generator.Generate()
	assert.NoError(t, err)

	assert.NoError(t, os.WriteFile("svg.gen.go.actual", actualSource, 0o666))

	expectedSource, err := os.ReadFile("svg.gen.go")
	assert.NoError(t, err)
	assert.Equal(t, string(expectedSource), string(actualSource))
}
