package aixm

import (
	"archive/zip"
	"encoding/xml"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"unicode"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/twpayne/go-xmlstruct"
)

var (
	exportRenames = map[string]string{
		"aixm":            "AIXM",
		"gco":             "GCO",
		"gmd":             "GMD",
		"gml":             "GML",
		"gsr":             "GSR",
		"gss":             "GSS",
		"gts":             "GTS",
		"id":              "ID",
		"ids":             "IDS",
		"note":            "LowerNote",            // Disambiguate between Note and note.
		"runwayDirection": "LowerRunwayDirection", // Disambiguate between RunwayDirection and runwayDirection.
		"uom":             "UOM",
		"xlink":           "XLink",
		"xsd":             "XSD",
		"xsi":             "XSI",
	}

	snakeCaseLetterRx = regexp.MustCompile(`_([A-Za-z])`)
)

func TestAIXM(t *testing.T) {
	generator := xmlstruct.NewGenerator(
		xmlstruct.WithExportNameFunc(func(name xml.Name) string {
			if exportName, ok := exportRenames[name.Local]; ok {
				return exportName
			}
			exportedName := name.Local
			exportedName = snakeCaseLetterRx.ReplaceAllStringFunc(exportedName, func(s string) string {
				return strings.ToUpper(s[1:])
			})
			runes := []rune(exportedName)
			runes[0] = unicode.ToUpper(runes[0])
			return string(runes)
		}),
		xmlstruct.WithNamedTypes(true),
		xmlstruct.WithPackageName("aixm"),
	)

	filenames := []string{
		"testdata/LF_AIP_DS_PartOf_20201203_AIRAC.zip",
		"testdata/LO_OBS_DS_AREA1_20221104_2022-10-25_1010984.zip",
	}

	observeZipFile := func(zipFile *zip.File) {
		readCloser, err := zipFile.Open()
		require.NoError(t, err)
		defer readCloser.Close()
		require.NoError(t, generator.ObserveReader(readCloser))
	}

	observeZipReader := func(zipReader *zip.Reader) {
		for _, zipFile := range zipReader.File {
			if filepath.Ext(zipFile.Name) == ".xml" {
				observeZipFile(zipFile)
			}
		}
	}

	zipReaders := make([]*zip.Reader, 0, len(filenames))
	for _, filename := range filenames {
		file, err := os.Open(filename)
		require.NoError(t, err)
		defer file.Close()

		fileInfo, err := file.Stat()
		require.NoError(t, err)

		zipReader, err := zip.NewReader(file, fileInfo.Size())
		require.NoError(t, err)

		observeZipReader(zipReader)

		zipReaders = append(zipReaders, zipReader)
	}

	actualSource, err := generator.Generate()
	require.NoError(t, err)

	require.NoError(t, os.WriteFile("aixm.gen.go.actual", actualSource, 0o666))

	expectedSource, err := os.ReadFile("aixm.gen.go")
	require.NoError(t, err)
	require.Equal(t, string(expectedSource), string(actualSource))

	decodeZipFile := func(zipFile *zip.File) {
		readCloser, err := zipFile.Open()
		require.NoError(t, err)
		defer readCloser.Close()

		var aixmBasicMessage AIXMBasicMessage
		require.NoError(t, xml.NewDecoder(readCloser).Decode(&aixmBasicMessage))

		switch zipFile.Name {
		case "LO_OBS_DS_AREA1_20221104.xml":
			assert.Equal(t, "https://sdimd-free.austrocontrol.at/geonetwork/srv/metadata/b0d38a5a-2072-42fc-8402-4ce984db8fae", aixmBasicMessage.MessageMetadata.MDMetadata.DataSetURI.CharacterString.CharData)
		case "LF_AIP_DS_PartOf_20201203_AIRAC.xml":
			assert.Equal(t, "uuid.729920d4-5360-49e3-b4b2-1a28313261ba", aixmBasicMessage.HasMember[0].AirportHeliport.ID)
		}
	}

	for _, zipReader := range zipReaders {
		for _, zipFile := range zipReader.File {
			if filepath.Ext(zipFile.Name) == ".xml" {
				decodeZipFile(zipFile)
			}
		}
	}
}
