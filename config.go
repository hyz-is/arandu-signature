package signature

import (
	"fmt"
	"net/url"

	"github.com/hyz-is/arandu-signature/internal/words"
)

// The two addresses the marks link to.
const (
	DefaultHyzisURL  = "https://hyz.is"
	DefaultAranduURL = "https://arandu.io"
)

// The languages the words of the line come in. Each is the tag of a language
// the line is written in; [Language] maps any page's tag to one of them.
const (
	// PortugueseBR reads "Desenvolvido por … com …", and is the default.
	PortugueseBR = words.PortugueseBR
	// English reads "Built by … with …", and is where an unknown language falls.
	English = words.English
	// French reads "Développé par … avec …".
	French = "fr"
	// Spanish reads "Desarrollado por … con …".
	Spanish = "es"
	// German reads "Entwickelt von … mit …".
	German = "de"
	// Italian reads "Sviluppato da … con …".
	Italian = "it"
	// Arabic reads right to left: the line is drawn with dir="rtl".
	Arabic = "ar"
	// Chinese is written in simplified characters ("zh-Hans").
	Chinese = "zh"
	// Japanese reads "… が … で開発".
	Japanese = "ja"
	// Russian reads "Разработано … на …".
	Russian = "ru"
)

// Language maps the language tag of a page ("fr", "pt_BR", "zh-Hans-CN",
// "ar-EG") to the language the line is drawn in. Case is ignored; Portuguese
// of any region is [PortugueseBR]; a traditional Chinese tag (script Hant, or
// region TW, HK or MO) is not supported, because the line is written in
// simplified characters. An empty tag is [PortugueseBR], and a tag the line
// has no words for is [English].
func Language(tag string) string { return words.Resolve(tag) }

// Languages lists the languages the line is written in.
func Languages() []string { return words.Tags() }

// Config is how a project signs.
//
// Enabled is the contract: the signature appears only on a product whose client
// agreed to the brand partnership, and a project that has not says nothing,
// so the line renders as nothing at all rather than as a hidden element.
type Config struct {
	// Enabled turns the signature on. Zero is off.
	Enabled bool
	// Lang is a language tag, such as a page's: see [Language]. Empty is
	// PortugueseBR and a tag the line has no words for is English. A tag that
	// is not shaped like a BCP 47 tag is refused by Validate.
	Lang string
	// HyzisURL and AranduURL are where the marks link to. Empty is the
	// default address of each; a value is refused unless it is an absolute
	// https address.
	HyzisURL  string
	AranduURL string

	// License is the partnership licence HYZIS issued for this project. With
	// a valid one for AppURL's host, the signature is always on: Enabled is
	// ignored, and Middleware puts the line back on any page that lost it.
	//
	// Keep it a Go constant in bootstrap/app.go rather than an environment
	// variable: removing it is then a change to the source, reviewed and
	// recorded in the repository like any other, not a setting flipped on a
	// server. An invalid or expired licence is logged and ignored; it never
	// stops the application.
	License string
	// AppURL is the address the application is served at (APP_URL), whose
	// host the licence has to name.
	AppURL string
}

// Validate refuses a configuration that would render a broken or unsafe link.
func (c Config) Validate() error {
	if c.Lang != "" && !words.WellFormed(c.Lang) {
		return fmt.Errorf("signature: Config.Lang is %q, and has to be a language tag such as %q or %q", c.Lang, PortugueseBR, English)
	}
	for name, address := range map[string]string{"HyzisURL": c.HyzisURL, "AranduURL": c.AranduURL} {
		if address == "" {
			continue
		}
		u, err := url.Parse(address)
		if err != nil || u.Scheme != "https" || u.Host == "" {
			return fmt.Errorf("signature: Config.%s is %q, and has to be an absolute https address", name, address)
		}
	}
	return nil
}

func (c Config) withDefaults() Config {
	if c.Lang == "" {
		c.Lang = PortugueseBR
	}
	if c.HyzisURL == "" {
		c.HyzisURL = DefaultHyzisURL
	}
	if c.AranduURL == "" {
		c.AranduURL = DefaultAranduURL
	}
	return c
}

// Words are the two pieces of text between the marks, in the language asked:
// before the HYZIS mark and before the Arandu mark. A language whose sentence
// ends after the second mark (Chinese, Japanese) has a third piece that Words
// does not carry; [Sentence] has all three.
func Words(lang string) (by, with string) {
	p := words.Lookup(lang)
	return p.Before, p.Between
}

// Phrase is the sentence around the two marks in one language.
type Phrase = words.Phrase

// Sentence answers the whole sentence for a language tag, as [Language]
// reads it: the text before, between and after the marks, whether the language
// is written right to left, and the tag for a lang attribute.
func Sentence(lang string) Phrase { return words.Lookup(lang) }
