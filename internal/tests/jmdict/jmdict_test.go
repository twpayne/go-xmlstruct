package jmdict_test

import (
	"compress/gzip"
	"encoding/xml"
	"errors"
	"io/fs"
	"os"
	"testing"

	"github.com/alecthomas/assert/v2"

	"github.com/twpayne/go-xmlstruct"
	"github.com/twpayne/go-xmlstruct/internal/tests/jmdict"
)

func TestJMDict(t *testing.T) {
	t.Parallel()

	entity := map[string]string{
		"bra":          "Brazilian",
		"hob":          "Hokkaido-ben",
		"ksb":          "Kansai-ben",
		"ktb":          "Kantou-ben",
		"kyb":          "Kyoto-ben",
		"kyu":          "Kyuushuu-ben",
		"nab":          "Nagano-ben",
		"osb":          "Osaka-ben",
		"rkb":          "Ryuukyuu-ben",
		"thb":          "Touhoku-ben",
		"tsb":          "Tosa-ben",
		"tsug":         "Tsugaru-ben",
		"agric":        "agriculture",
		"anat":         "anatomy",
		"archeol":      "archeology",
		"archit":       "architecture",
		"art":          "art, aesthetics",
		"astron":       "astronomy",
		"audvid":       "audiovisual",
		"aviat":        "aviation",
		"baseb":        "baseball",
		"biochem":      "biochemistry",
		"biol":         "biology",
		"bot":          "botany",
		"boxing":       "boxing",
		"Buddh":        "Buddhism",
		"bus":          "business",
		"cards":        "card games",
		"chem":         "chemistry",
		"chmyth":       "Chinese mythology",
		"Christn":      "Christianity",
		"civeng":       "civil engineering",
		"cloth":        "clothing",
		"comp":         "computing",
		"cryst":        "crystallography",
		"dent":         "dentistry",
		"ecol":         "ecology",
		"econ":         "economics",
		"elec":         "electricity, elec. eng.",
		"electr":       "electronics",
		"embryo":       "embryology",
		"engr":         "engineering",
		"ent":          "entomology",
		"figskt":       "figure skating",
		"film":         "film",
		"finc":         "finance",
		"fish":         "fishing",
		"food":         "food, cooking",
		"gardn":        "gardening, horticulture",
		"genet":        "genetics",
		"geogr":        "geography",
		"geol":         "geology",
		"geom":         "geometry",
		"go":           "go (game)",
		"golf":         "golf",
		"gramm":        "grammar",
		"grmyth":       "Greek mythology",
		"hanaf":        "hanafuda",
		"horse":        "horse racing",
		"internet":     "Internet",
		"jpmyth":       "Japanese mythology",
		"kabuki":       "kabuki",
		"law":          "law",
		"ling":         "linguistics",
		"logic":        "logic",
		"MA":           "martial arts",
		"mahj":         "mahjong",
		"manga":        "manga",
		"math":         "mathematics",
		"mech":         "mechanical engineering",
		"med":          "medicine",
		"met":          "meteorology",
		"mil":          "military",
		"min":          "mineralogy",
		"mining":       "mining",
		"motor":        "motorsport",
		"music":        "music",
		"noh":          "noh",
		"ornith":       "ornithology",
		"paleo":        "paleontology",
		"pathol":       "pathology",
		"pharm":        "pharmacology",
		"phil":         "philosophy",
		"photo":        "photography",
		"physics":      "physics",
		"physiol":      "physiology",
		"politics":     "politics",
		"print":        "printing",
		"prowres":      "professional wrestling",
		"psy":          "psychiatry",
		"psyanal":      "psychoanalysis",
		"psych":        "psychology",
		"rail":         "railway",
		"rommyth":      "Roman mythology",
		"Shinto":       "Shinto",
		"shogi":        "shogi",
		"ski":          "skiing",
		"sports":       "sports",
		"stat":         "statistics",
		"stockm":       "stock market",
		"sumo":         "sumo",
		"surg":         "surgery",
		"telec":        "telecommunications",
		"tradem":       "trademark",
		"tv":           "television",
		"vet":          "veterinary terms",
		"vidg":         "video games",
		"zool":         "zoology",
		"ateji":        "ateji (phonetic) reading",
		"ik":           "word containing irregular kana usage",
		"iK":           "word containing irregular kanji usage",
		"io":           "irregular okurigana usage",
		"oK":           "word containing out-dated kanji or kanji usage",
		"rK":           "rarely used kanji form",
		"sK":           "search-only kanji form",
		"abbr":         "abbreviation",
		"arch":         "archaic",
		"char":         "character",
		"chn":          "children's language",
		"col":          "colloquial",
		"company":      "company name",
		"creat":        "creature",
		"dated":        "dated term",
		"dei":          "deity",
		"derog":        "derogatory",
		"doc":          "document",
		"euph":         "euphemistic",
		"ev":           "event",
		"fam":          "familiar language",
		"fem":          "female term or language",
		"fict":         "fiction",
		"form":         "formal or literary term",
		"given":        "given name or forename, gender not specified",
		"group":        "group",
		"hist":         "historical term",
		"hon":          "honorific or respectful (sonkeigo) language",
		"hum":          "humble (kenjougo) language",
		"id":           "idiomatic expression",
		"joc":          "jocular, humorous term",
		"leg":          "legend",
		"m-sl":         "manga slang",
		"male":         "male term or language",
		"myth":         "mythology",
		"net-sl":       "Internet slang",
		"obj":          "object",
		"obs":          "obsolete term",
		"on-mim":       "onomatopoeic or mimetic word",
		"organization": "organization name",
		"oth":          "other",
		"person":       "full name of a particular person",
		"place":        "place name",
		"poet":         "poetical term",
		"pol":          "polite (teineigo) language",
		"product":      "product name",
		"proverb":      "proverb",
		"quote":        "quotation",
		"rare":         "rare term",
		"relig":        "religion",
		"sens":         "sensitive",
		"serv":         "service",
		"ship":         "ship name",
		"sl":           "slang",
		"station":      "railway station",
		"surname":      "family or surname",
		"uk":           "word usually written using kana alone",
		"unclass":      "unclassified name",
		"vulg":         "vulgar expression or word",
		"work":         "work of art, literature, music, etc. name",
		"X":            "rude or X-rated term (not displayed in educational software)",
		"yoji":         "yojijukugo",
		"adj-f":        "noun or verb acting prenominally",
		"adj-i":        "adjective (keiyoushi)",
		"adj-ix":       "adjective (keiyoushi) - yoi/ii class",
		"adj-kari":     "'kari' adjective (archaic)",
		"adj-ku":       "'ku' adjective (archaic)",
		"adj-na":       "adjectival nouns or quasi-adjectives (keiyodoshi)",
		"adj-nari":     "archaic/formal form of na-adjective",
		"adj-no":       "nouns which may take the genitive case particle 'no'",
		"adj-pn":       "pre-noun adjectival (rentaishi)",
		"adj-shiku":    "'shiku' adjective (archaic)",
		"adj-t":        "'taru' adjective",
		"adv":          "adverb (fukushi)",
		"adv-to":       "adverb taking the 'to' particle",
		"aux":          "auxiliary",
		"aux-adj":      "auxiliary adjective",
		"aux-v":        "auxiliary verb",
		"conj":         "conjunction",
		"cop":          "copula",
		"ctr":          "counter",
		"exp":          "expressions (phrases, clauses, etc.)",
		"int":          "interjection (kandoushi)",
		"n":            "noun (common) (futsuumeishi)",
		"n-adv":        "adverbial noun (fukushitekimeishi)",
		"n-pr":         "proper noun",
		"n-pref":       "noun, used as a prefix",
		"n-suf":        "noun, used as a suffix",
		"n-t":          "noun (temporal) (jisoumeishi)",
		"num":          "numeric",
		"pn":           "pronoun",
		"pref":         "prefix",
		"prt":          "particle",
		"suf":          "suffix",
		"unc":          "unclassified",
		"v-unspec":     "verb unspecified",
		"v1":           "Ichidan verb",
		"v1-s":         "Ichidan verb - kureru special class",
		"v2a-s":        "Nidan verb with 'u' ending (archaic)",
		"v2b-k":        "Nidan verb (upper class) with 'bu' ending (archaic)",
		"v2b-s":        "Nidan verb (lower class) with 'bu' ending (archaic)",
		"v2d-k":        "Nidan verb (upper class) with 'dzu' ending (archaic)",
		"v2d-s":        "Nidan verb (lower class) with 'dzu' ending (archaic)",
		"v2g-k":        "Nidan verb (upper class) with 'gu' ending (archaic)",
		"v2g-s":        "Nidan verb (lower class) with 'gu' ending (archaic)",
		"v2h-k":        "Nidan verb (upper class) with 'hu/fu' ending (archaic)",
		"v2h-s":        "Nidan verb (lower class) with 'hu/fu' ending (archaic)",
		"v2k-k":        "Nidan verb (upper class) with 'ku' ending (archaic)",
		"v2k-s":        "Nidan verb (lower class) with 'ku' ending (archaic)",
		"v2m-k":        "Nidan verb (upper class) with 'mu' ending (archaic)",
		"v2m-s":        "Nidan verb (lower class) with 'mu' ending (archaic)",
		"v2n-s":        "Nidan verb (lower class) with 'nu' ending (archaic)",
		"v2r-k":        "Nidan verb (upper class) with 'ru' ending (archaic)",
		"v2r-s":        "Nidan verb (lower class) with 'ru' ending (archaic)",
		"v2s-s":        "Nidan verb (lower class) with 'su' ending (archaic)",
		"v2t-k":        "Nidan verb (upper class) with 'tsu' ending (archaic)",
		"v2t-s":        "Nidan verb (lower class) with 'tsu' ending (archaic)",
		"v2w-s":        "Nidan verb (lower class) with 'u' ending and 'we' conjugation (archaic)",
		"v2y-k":        "Nidan verb (upper class) with 'yu' ending (archaic)",
		"v2y-s":        "Nidan verb (lower class) with 'yu' ending (archaic)",
		"v2z-s":        "Nidan verb (lower class) with 'zu' ending (archaic)",
		"v4b":          "Yodan verb with 'bu' ending (archaic)",
		"v4g":          "Yodan verb with 'gu' ending (archaic)",
		"v4h":          "Yodan verb with 'hu/fu' ending (archaic)",
		"v4k":          "Yodan verb with 'ku' ending (archaic)",
		"v4m":          "Yodan verb with 'mu' ending (archaic)",
		"v4n":          "Yodan verb with 'nu' ending (archaic)",
		"v4r":          "Yodan verb with 'ru' ending (archaic)",
		"v4s":          "Yodan verb with 'su' ending (archaic)",
		"v4t":          "Yodan verb with 'tsu' ending (archaic)",
		"v5aru":        "Godan verb - -aru special class",
		"v5b":          "Godan verb with 'bu' ending",
		"v5g":          "Godan verb with 'gu' ending",
		"v5k":          "Godan verb with 'ku' ending",
		"v5k-s":        "Godan verb - Iku/Yuku special class",
		"v5m":          "Godan verb with 'mu' ending",
		"v5n":          "Godan verb with 'nu' ending",
		"v5r":          "Godan verb with 'ru' ending",
		"v5r-i":        "Godan verb with 'ru' ending (irregular verb)",
		"v5s":          "Godan verb with 'su' ending",
		"v5t":          "Godan verb with 'tsu' ending",
		"v5u":          "Godan verb with 'u' ending",
		"v5u-s":        "Godan verb with 'u' ending (special class)",
		"v5uru":        "Godan verb - Uru old class verb (old form of Eru)",
		"vi":           "intransitive verb",
		"vk":           "Kuru verb - special class",
		"vn":           "irregular nu verb",
		"vr":           "irregular ru verb, plain form ends with -ri",
		"vs":           "noun or participle which takes the aux. verb suru",
		"vs-c":         "su verb - precursor to the modern suru",
		"vs-i":         "suru verb - included",
		"vs-s":         "suru verb - special class",
		"vt":           "transitive verb",
		"vz":           "Ichidan verb - zuru verb (alternative form of -jiru verbs)",
		"gikun":        "gikun (meaning as reading) or jukujikun (special kanji reading)",
		"ok":           "out-dated or obsolete kana usage",
		"rk":           "rarely used kana form",
		"sk":           "search-only kana form",
	}

	generator := xmlstruct.NewGenerator(
		xmlstruct.WithExportRenames(map[string]string{
			"JMdict":     "JMDict",
			"ls_type":    "LSType",
			"ls_wasei":   "LSWasei",
			"lsource":    "LSource",
			"re_nokanji": "ReNoKanji",
			"stagk":      "STagK",
			"stagr":      "STagR",
			"xref":       "XRef",
		}),
		xmlstruct.WithModifyDecoderFunc(func(d *xml.Decoder) {
			d.Entity = entity
		}),
		xmlstruct.WithNamedTypes(true),
		xmlstruct.WithPackageName("jmdict"),
	)

	func() {
		file, err := os.Open("testdata/JMdict.gz")
		if errors.Is(err, fs.ErrNotExist) {
			t.Skip("missing testdata/JMdict.gz")
		}
		assert.NoError(t, err)
		defer file.Close()
		gzipReader, err := gzip.NewReader(file)
		assert.NoError(t, err)
		defer gzipReader.Close()
		assert.NoError(t, generator.ObserveReader(gzipReader))
	}()

	actualSource, err := generator.Generate()
	assert.NoError(t, err)

	assert.NoError(t, os.WriteFile("jmdict.gen.go.actual", actualSource, 0o666))

	expectedSource, err := os.ReadFile("jmdict.gen.go")
	assert.NoError(t, err)
	assert.Equal(t, string(expectedSource), string(actualSource))

	var jmDict jmdict.JMDict
	func() {
		file, err := os.Open("testdata/JMdict.gz")
		if errors.Is(err, fs.ErrNotExist) {
			t.Skip("missing testdata/JMdict.gz")
		}
		assert.NoError(t, err)
		defer file.Close()
		gzipReader, err := gzip.NewReader(file)
		assert.NoError(t, err)
		defer gzipReader.Close()

		decoder := xml.NewDecoder(gzipReader)
		decoder.Entity = entity
		assert.NoError(t, decoder.Decode(&jmDict))
	}()

	assert.Equal(t, "repetition mark in katakana", jmDict.Entry[0].Sense[0].Gloss[0].CharData)
}
