package play

import (
	"encoding/xml"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/twpayne/go-xmlstruct"
)

func TestPlay(t *testing.T) {
	generator := xmlstruct.NewGenerator(
		xmlstruct.WithExportRenames(map[string]string{
			"ACT":      "Act",
			"EPILOGUE": "Epilogue",
			"FM":       "FrontMatter",
			"GRPDESCR": "GroupDescription",
			"LINE":     "Line",
			"P":        "Paragraph",
			"PERSONA":  "Persona",
			"PERSONAE": "Personae",
			"PGROUP":   "PersonaGroup",
			"PLAY":     "Play",
			"PLAYSUBT": "PlaySubtitle",
			"SCENE":    "Scene",
			"SCNDESCR": "SceneDescription",
			"SPEAKER":  "Speaker",
			"SPEECH":   "Speech",
			"STAGEDIR": "StageDirection",
			"TITLE":    "Title",
		}),
		xmlstruct.WithNamedTypes(true),
		xmlstruct.WithPackageName("play"),
		xmlstruct.WithPreserveOrder(true),
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
