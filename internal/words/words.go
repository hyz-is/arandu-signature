// Package words holds the sentence of the signature in every language it
// speaks, so the web, e-mail and native halves read one table.
package words

import "strings"

// Phrase is the sentence around the two marks: Before, the HYZIS mark,
// Between, the Arandu mark, After. A language whose verb comes last (Chinese,
// Japanese) leaves Before empty or fills After; one that reads right to left
// sets RTL and is drawn with dir="rtl", which reverses the row.
type Phrase struct {
	// Tag is the supported language the phrase is written in, as its constant.
	Tag string
	// HTMLLang is the BCP 47 tag for the lang attribute.
	HTMLLang string
	// Before, Between and After are the pieces of text; any may be empty.
	Before, Between, After string
	// Version is the word for "Version" on an About screen.
	Version string
	// RTL is true for a language written right to left.
	RTL bool
}

// Dir is the value of the dir attribute.
func (p Phrase) Dir() string {
	if p.RTL {
		return "rtl"
	}
	return "ltr"
}

const (
	// PortugueseBR is the default of an empty tag.
	PortugueseBR = "pt-BR"
	// English is where a language without a phrase falls to.
	English = "en"
)

var table = []Phrase{
	{Tag: "fr", HTMLLang: "fr", Before: "Développé par", Between: "avec", Version: "Version"},
	{Tag: English, HTMLLang: "en", Before: "Built by", Between: "with", Version: "Version"},
	{Tag: "es", HTMLLang: "es", Before: "Desarrollado por", Between: "con", Version: "Versión"},
	{Tag: PortugueseBR, HTMLLang: "pt-BR", Before: "Desenvolvido por", Between: "com", Version: "Versão"},
	{Tag: "de", HTMLLang: "de", Before: "Entwickelt von", Between: "mit", Version: "Version"},
	{Tag: "it", HTMLLang: "it", Before: "Sviluppato da", Between: "con", Version: "Versione"},
	{Tag: "ar", HTMLLang: "ar", Before: "تم التطوير بواسطة", Between: "باستخدام", Version: "الإصدار", RTL: true},
	{Tag: "zh", HTMLLang: "zh-Hans", Before: "由", Between: "使用", After: "开发", Version: "版本"},
	{Tag: "ja", HTMLLang: "ja", Between: "が", After: "で開発", Version: "バージョン"},
	{Tag: "ru", HTMLLang: "ru", Before: "Разработано", Between: "на", Version: "Версия"},
}

// Tags lists the supported languages, in the order of the table.
func Tags() []string {
	out := make([]string, len(table))
	for i, p := range table {
		out[i] = p.Tag
	}
	return out
}

// Lookup answers the phrase for a language tag. An empty tag is Portuguese
// (Brazil); a tag that is not supported is English.
func Lookup(tag string) Phrase {
	want := Resolve(tag)
	for _, p := range table {
		if p.Tag == want {
			return p
		}
	}
	return table[1]
}

// Resolve maps a page's BCP 47 tag ("fr", "pt_BR", "zh-Hans-CN", "AR-EG") to
// the supported language it is drawn in. Case is ignored and "_" counts as
// "-". Portuguese of any region is "pt-BR". Chinese is written in simplified
// characters, so a tag that says traditional (script Hant, or region TW, HK or
// MO) is not Chinese here. Empty is "pt-BR"; anything else is "en".
func Resolve(tag string) string {
	tag = strings.ToLower(strings.TrimSpace(strings.ReplaceAll(tag, "_", "-")))
	if tag == "" {
		return PortugueseBR
	}
	parts := strings.Split(tag, "-")
	switch lang := parts[0]; lang {
	case "pt":
		return PortugueseBR
	case "zh":
		for _, sub := range parts[1:] {
			switch sub {
			case "hant", "tw", "hk", "mo":
				return English
			}
		}
		return "zh"
	case "fr", "en", "es", "de", "it", "ar", "ja", "ru":
		return lang
	}
	return English
}

// WellFormed reports whether a tag is shaped like a BCP 47 tag: subtags of one
// to eight letters or digits, the first of two or three letters.
func WellFormed(tag string) bool {
	parts := strings.Split(strings.ReplaceAll(tag, "_", "-"), "-")
	if n := len(parts[0]); n < 2 || n > 3 {
		return false
	}
	for i, p := range parts {
		if len(p) == 0 || len(p) > 8 {
			return false
		}
		for _, r := range p {
			letter := (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
			if !letter && (i == 0 || r < '0' || r > '9') {
				return false
			}
		}
	}
	return true
}
