package unit_test

import (
	"strings"
	"testing"

	"github.com/hyz-is/arandu-signature"
)

// Each language the line speaks, with the text before the HYZIS mark, between
// the marks and after the Arandu mark.
var languages = []struct {
	tag, htmlLang, before, between, after, dir string
}{
	{signature.French, "fr", "Développé par", "avec", "", "ltr"},
	{signature.English, "en", "Built by", "with", "", "ltr"},
	{signature.Spanish, "es", "Desarrollado por", "con", "", "ltr"},
	{signature.PortugueseBR, "pt-BR", "Desenvolvido por", "com", "", "ltr"},
	{signature.German, "de", "Entwickelt von", "mit", "", "ltr"},
	{signature.Italian, "it", "Sviluppato da", "con", "", "ltr"},
	{signature.Arabic, "ar", "تم التطوير بواسطة", "باستخدام", "", "rtl"},
	{signature.Chinese, "zh-Hans", "由", "使用", "开发", "ltr"},
	{signature.Japanese, "ja", "", "が", "で開発", "ltr"},
	{signature.Russian, "ru", "Разработано", "на", "", "ltr"},
}

func TestTheLineIsDrawnInEveryLanguage(t *testing.T) {
	t.Parallel()

	if got := len(signature.Languages()); got != len(languages) {
		t.Fatalf("the module lists %d languages, the test covers %d", got, len(languages))
	}
	for _, l := range languages {
		t.Run(l.tag, func(t *testing.T) {
			html := string(signature.Kyse(signature.Props{Enabled: true, Lang: l.tag}))
			for _, want := range []string{`lang="` + l.htmlLang + `"`, `dir="` + l.dir + `"`} {
				if !strings.Contains(html, want) {
					t.Errorf("the line lacks %s:\n%s", want, html)
				}
			}
			// The pieces appear in this order: before, HYZIS, between, Arandu, after.
			order := []string{`aria-label="HYZIS"`, `aria-label="Arandu"`}
			if l.before != "" {
				order = append([]string{"<span>" + l.before + "</span>"}, order...)
			}
			order = append(order[:len(order)-1], append([]string{"<span>" + l.between + "</span>"}, order[len(order)-1])...)
			if l.after != "" {
				order = append(order, "<span>"+l.after+"</span>")
			}
			at := 0
			for _, piece := range order {
				i := strings.Index(html[at:], piece)
				if i < 0 {
					t.Fatalf("%s is missing or out of order in:\n%s", piece, html)
				}
				at += i + len(piece)
			}
			if l.before == "" && strings.Contains(html, "<span></span>") {
				t.Errorf("the line draws an empty span:\n%s", html)
			}

			before, between := signature.Words(l.tag)
			if before != l.before || between != l.between {
				t.Errorf("Words(%q) = %q, %q; want %q, %q", l.tag, before, between, l.before, l.between)
			}
			mail := string(signature.Email(signature.EmailProps{Enabled: true, Origin: letterOrigin, Lang: l.tag}))
			for _, want := range []string{`lang="` + l.htmlLang + `"`, `dir="` + l.dir + `"`, ">" + l.between + "</td>"} {
				if !strings.Contains(mail, want) {
					t.Errorf("the e-mail line lacks %s:\n%s", want, mail)
				}
			}
			if l.after != "" && !strings.Contains(mail, ">"+l.after+"</td>") {
				t.Errorf("the e-mail line lacks %s:\n%s", l.after, mail)
			}
			text := signature.EmailText(signature.EmailProps{Enabled: true, Lang: l.tag})
			want := strings.TrimSpace(l.before + " HYZIS (https://hyz.is) " + l.between + " Arandu (https://arandu.io) " + l.after)
			if text != want {
				t.Errorf("text = %q, want %q", text, want)
			}
		})
	}
}

func TestArabicIsDrawnRightToLeft(t *testing.T) {
	t.Parallel()

	for _, tag := range []string{"ar", "ar-EG", "AR_sa"} {
		if html := string(signature.Kyse(signature.Props{Enabled: true, Lang: tag})); !strings.Contains(html, `dir="rtl"`) || !strings.Contains(html, `lang="ar"`) {
			t.Errorf("%s is not drawn right to left:\n%s", tag, html)
		}
		if mail := string(signature.Email(signature.EmailProps{Enabled: true, Origin: letterOrigin, Lang: tag})); !strings.Contains(mail, `dir="rtl"`) {
			t.Errorf("%s e-mail is not right to left:\n%s", tag, mail)
		}
	}
	if !signature.Sentence("ar").RTL || signature.Sentence("fr").RTL {
		t.Error("only Arabic is right to left")
	}
}

func TestAnUnknownLanguageIsEnglish(t *testing.T) {
	t.Parallel()

	for _, tag := range []string{"ko", "xx", "hi-IN", "zh-TW", "zh-Hant", "zh-HK", "sw"} {
		if got := signature.Language(tag); got != signature.English {
			t.Errorf("Language(%q) = %q, want %q", tag, got, signature.English)
		}
		html := string(signature.Kyse(signature.Props{Enabled: true, Lang: tag}))
		if !strings.Contains(html, "<span>Built by</span>") || !strings.Contains(html, `lang="en"`) {
			t.Errorf("%s does not fall to English:\n%s", tag, html)
		}
	}
	// A tag that is not shaped like one does not stop the line either.
	if html := string(signature.Kyse(signature.Props{Enabled: true, Lang: "!!"})); !strings.Contains(html, "<span>Built by</span>") {
		t.Errorf("a malformed tag does not fall to English:\n%s", html)
	}
}

func TestAPageTagIsMappedToTheLanguage(t *testing.T) {
	t.Parallel()

	for tag, want := range map[string]string{
		"":           signature.PortugueseBR,
		"pt":         signature.PortugueseBR,
		"pt-BR":      signature.PortugueseBR,
		"pt_PT":      signature.PortugueseBR,
		"PT-br":      signature.PortugueseBR,
		"fr":         signature.French,
		"fr-CA":      signature.French,
		"es-419":     signature.Spanish,
		"de-AT":      signature.German,
		"it":         signature.Italian,
		"ar":         signature.Arabic,
		"zh":         signature.Chinese,
		"zh-CN":      signature.Chinese,
		"zh-Hans":    signature.Chinese,
		"zh-Hans-CN": signature.Chinese,
		"ja-JP":      signature.Japanese,
		"ru":         signature.Russian,
		"en-GB":      signature.English,
	} {
		if got := signature.Language(tag); got != want {
			t.Errorf("Language(%q) = %q, want %q", tag, got, want)
		}
	}
}

func TestTheConfigurationAcceptsEveryLanguageAndKeepsEnglishAndPortuguese(t *testing.T) {
	t.Parallel()

	for _, l := range languages {
		m, err := signature.New(signature.Config{Enabled: true, Lang: l.tag})
		if err != nil {
			t.Fatalf("New refused %q: %v", l.tag, err)
		}
		if !strings.Contains(string(m.HTML()), `lang="`+l.htmlLang+`"`) {
			t.Errorf("%s: the module draws %q", l.tag, m.HTML())
		}
	}
	if _, err := signature.New(signature.Config{Enabled: true, Lang: "ko"}); err != nil {
		t.Errorf("a language without words is English, not an error: %v", err)
	}
}
