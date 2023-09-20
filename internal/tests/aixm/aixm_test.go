package aixm

import (
	"archive/zip"
	"encoding/xml"
	"os"
	"path/filepath"
	"testing"

	"github.com/alecthomas/assert/v2"

	"github.com/twpayne/go-xmlstruct"
)

func TestAIXM(t *testing.T) {
	t.Parallel()

	generator := xmlstruct.NewGenerator(
		xmlstruct.WithExportRenames(map[string]string{
			"note":            "NoteLower",            // Disambiguate between Note and note.
			"runwayDirection": "RunwayDirectionLower", // Disambiguate between RunwayDirection and runwayDirection.
			"uom":             "UOM",                  // Capitalize unit of measurement abbreviation.
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
		assert.NoError(t, err)
		defer readCloser.Close()
		assert.NoError(t, generator.ObserveReader(readCloser))
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
		assert.NoError(t, err)
		defer file.Close()

		fileInfo, err := file.Stat()
		assert.NoError(t, err)

		zipReader, err := zip.NewReader(file, fileInfo.Size())
		assert.NoError(t, err)

		observeZipReader(zipReader)

		zipReaders = append(zipReaders, zipReader)
	}

	actualSource, err := generator.Generate()
	assert.NoError(t, err)

	assert.NoError(t, os.WriteFile("aixm.gen.go.actual", actualSource, 0o666))

	expectedSource, err := os.ReadFile("aixm.gen.go")
	assert.NoError(t, err)
	assert.Equal(t, string(expectedSource), string(actualSource))

	decodeZipFile := func(zipFile *zip.File) *AIXMBasicMessage {
		readCloser, err := zipFile.Open()
		assert.NoError(t, err)
		defer readCloser.Close()

		var aixmBasicMessage AIXMBasicMessage
		assert.NoError(t, xml.NewDecoder(readCloser).Decode(&aixmBasicMessage))
		return &aixmBasicMessage
	}

	for _, zipReader := range zipReaders {
		for _, zipFile := range zipReader.File {
			if filepath.Ext(zipFile.Name) == ".xml" {
				aixmBasicMessage := decodeZipFile(zipFile)
				switch zipFile.Name {
				case "LO_OBS_DS_AREA1_20221104.xml":
					assert.Equal(t, "https://sdimd-free.austrocontrol.at/geonetwork/srv/metadata/b0d38a5a-2072-42fc-8402-4ce984db8fae", aixmBasicMessage.MessageMetadata.MDMetadata.DataSetURI.CharacterString)
				case "LF_AIP_DS_PartOf_20201203_AIRAC.xml":
					assert.Equal(t, "uuid.729920d4-5360-49e3-b4b2-1a28313261ba", aixmBasicMessage.HasMember[0].AirportHeliport.ID)
				}
			}
		}
	}
}
