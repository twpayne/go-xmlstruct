package play

import (
	"encoding/xml"
	"os"
	"testing"

	"github.com/alecthomas/assert/v2"

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

	assert.NoError(t, generator.ObserveFile("testdata/all_well.xml"))

	actualSource, err := generator.Generate()
	assert.NoError(t, err)

	assert.NoError(t, os.WriteFile("play.gen.go.actual", actualSource, 0o666))

	expectedSource, err := os.ReadFile("play.gen.go")
	assert.NoError(t, err)
	assert.Equal(t, string(expectedSource), string(actualSource))

	data, err := os.ReadFile("testdata/all_well.xml")
	assert.NoError(t, err)
	var allsWellThatEndsWell Play
	assert.NoError(t, xml.Unmarshal(data, &allsWellThatEndsWell))

	assert.Equal(t, 5, len(allsWellThatEndsWell.Act))
	assert.Equal(t, "All's well that ends well; still the fine's the crown;", allsWellThatEndsWell.Act[3].Scene[3].Speech[4].Line[5].CharData)
}
