package play

import (
	"encoding/xml"
	"os"
	"strings"
	"testing"
	"unicode"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/twpayne/go-xmlstruct"
)

var exportRenames = map[string]string{
	"FM":       "FrontMatter",
	"GRPDESCR": "GroupDescription",
	"P":        "Paragraph",
	"PGROUP":   "PersonaGroup",
	"PLAYSUBT": "PlaySubtitle",
	"SCNDESCR": "ScreenDescription",
	"STAGEDIR": "StageDirection",
}

func TestPlay(t *testing.T) {
	generator := xmlstruct.NewGenerator(
		xmlstruct.WithExportNameFunc(func(name xml.Name) string {
			if exportName, ok := exportRenames[name.Local]; ok {
				return exportName
			}
			runes := []rune(strings.ToLower(name.Local))
			runes[0] = unicode.ToUpper(runes[0])
			return string(runes)
		}),
		xmlstruct.WithNamedTypes(true),
		xmlstruct.WithPackageName("play"),
	)

	require.NoError(t, generator.ObserveFile("testdata/all_well.xml"))

	actualSource, err := generator.Generate()
	require.NoError(t, err)

	require.NoError(t, os.WriteFile("play.gen.go.actual", actualSource, 0o666))

	expectedSource, err := os.ReadFile("play.gen.go")
	require.NoError(t, err)
	require.Equal(t, string(expectedSource), string(actualSource))

	data, err := os.ReadFile("testdata/all_well.xml")
	require.NoError(t, err)
	var allsWellThatEndsWell Play
	require.NoError(t, xml.Unmarshal(data, &allsWellThatEndsWell))

	assert.Len(t, allsWellThatEndsWell.Act, 5)
	assert.Equal(t, "All's well that ends well; still the fine's the crown;", allsWellThatEndsWell.Act[3].Scene[3].Speech[4].Line[5].CharData)
}
