package aixm

import (
	"archive/zip"
	"encoding/xml"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/twpayne/go-xmlstruct"
)

func TestAIXM(t *testing.T) {
	file, err := os.Open("testdata/LF_AIP_DS_PartOf_20201203_AIRAC.zip")
	require.NoError(t, err)
	defer file.Close()

	fileInfo, err := file.Stat()
	require.NoError(t, err)

	zipReader, err := zip.NewReader(file, fileInfo.Size())
	require.NoError(t, err)

	generator := xmlstruct.NewGenerator(
		xmlstruct.WithPackageName("aixm"),
	)
	observeZipFile := func(zipFile *zip.File) {
		readCloser, err := zipFile.Open()
		require.NoError(t, err)
		defer readCloser.Close()
		require.NoError(t, generator.ObserveReader(readCloser))
	}
	for _, zipFile := range zipReader.File {
		observeZipFile(zipFile)
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
		case "LF_AIP_DS_PartOf_20201203_AIRAC.xml":
			assert.Equal(t, "uuid.729920d4-5360-49e3-b4b2-1a28313261ba", aixmBasicMessage.HasMember[0].AirportHeliport.Id)
		}
	}
	for _, zipFile := range zipReader.File {
		decodeZipFile(zipFile)
	}
}
