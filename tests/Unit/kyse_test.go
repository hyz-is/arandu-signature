package unit_test

import (
	"strings"
	"testing"

	"github.com/arandu-io/hesape/view"

	signature "github.com/hyz-is/arandu-signature"
)

// A project whose client did not agree to the partnership signs nothing: not a
// hidden element, nothing at all.
func TestTheLineIsNothingUntilTheContractAllowsIt(t *testing.T) {
	t.Parallel()

	if got := signature.Kyse(signature.Props{}); got != "" {
		t.Fatalf("a disabled signature rendered %q", got)
	}
}

// The line reads as a sentence, each mark a link named for a screen reader and
// pointing at its site, with its stylesheet served from the origin.
func TestTheLineIsTheFamilySignature(t *testing.T) {
	t.Parallel()

	html := string(signature.Kyse(signature.Props{Enabled: true, Class: "footer-credit"}))
	for _, want := range []string{
		`<span>Desenvolvido por</span>`,
		`<span>em</span>`,
		`href="https://hyz.is" aria-label="HYZIS"`,
		`href="https://arandu.io" aria-label="Arandu"`,
		`class="arandu-signature footer-credit"`,
		`lang="pt-BR"`,
		`aria-hidden="true"`,
		`rel="stylesheet" href="` + view.Asset(signature.StylesheetAsset) + `"`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("the line lacks %s:\n%s", want, html)
		}
	}
	for _, refused := range []string{"style=", "<script", "onclick"} {
		if strings.Contains(html, refused) {
			t.Errorf("the line carries %s, which a strict Content-Security-Policy refuses:\n%s", refused, html)
		}
	}
}

func TestTheLineSpeaksEnglishWhenAsked(t *testing.T) {
	t.Parallel()

	html := string(signature.Kyse(signature.Props{Enabled: true, Lang: signature.English}))
	if !strings.Contains(html, "<span>Built by</span>") || !strings.Contains(html, "<span>with</span>") || !strings.Contains(html, `lang="en"`) {
		t.Fatalf("the English line reads:\n%s", html)
	}
}

func TestTheLinksCanPointElsewhereOnlyOverHTTPS(t *testing.T) {
	t.Parallel()

	html := string(signature.Kyse(signature.Props{Enabled: true, HyzisURL: "https://hyz.is/cases"}))
	if !strings.Contains(html, `href="https://hyz.is/cases"`) {
		t.Fatalf("the HYZIS link was not overridden:\n%s", html)
	}
	if got := signature.Kyse(signature.Props{Enabled: true, AranduURL: "javascript:alert(1)"}); got != "" {
		t.Fatalf("an unsafe address rendered %q", got)
	}
}

func TestTheConfigurationRefusesWhatCannotWork(t *testing.T) {
	t.Parallel()

	for name, cfg := range map[string]signature.Config{
		"unknown language": {Lang: "fr"},
		"plain http":       {HyzisURL: "http://hyz.is"},
		"relative address": {AranduURL: "/arandu"},
	} {
		if _, err := signature.New(cfg); err == nil {
			t.Errorf("%s: New accepted %+v", name, cfg)
		}
	}
	m, err := signature.New(signature.Config{Enabled: true})
	if err != nil {
		t.Fatalf("New refused the default configuration: %v", err)
	}
	if !m.Enabled() || !strings.Contains(string(m.HTML()), "Desenvolvido por") {
		t.Fatalf("the module does not render its line: %q", m.HTML())
	}
}

// The marks and the stylesheet are registered as view assets, each served
// under its own hashed address.
func TestTheAssetsAreRegistered(t *testing.T) {
	t.Parallel()

	registered := map[string]string{}
	for _, a := range view.Assets() {
		registered[a.Name] = a.ContentType
	}
	for name, want := range map[string]string{
		signature.StylesheetAsset: "text/css",
		signature.HyzisAsset:      "image/svg+xml",
		signature.AranduAsset:     "image/svg+xml",
	} {
		got, ok := registered[name]
		if !ok {
			t.Errorf("%s is not a registered asset", name)
			continue
		}
		if !strings.HasPrefix(got, want) {
			t.Errorf("%s is served as %q, want %s", name, got, want)
		}
		if !strings.Contains(view.Asset(name), name) {
			t.Errorf("%s is served at %q", name, view.Asset(name))
		}
	}
}
