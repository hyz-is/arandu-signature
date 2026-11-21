package unit_test

import (
	"bytes"
	"image/png"
	"math"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/arandu-io/hesape/view"

	signature "github.com/hyz-is/arandu-signature"
)

const letterOrigin = "https://letters.example.test"

// A project whose client did not agree to the partnership signs nothing in an
// e-mail either: no table, no images.
func TestTheEmailLineIsNothingUntilTheContractAllowsIt(t *testing.T) {
	t.Parallel()

	if got := signature.Email(signature.EmailProps{Origin: letterOrigin}); got != "" {
		t.Fatalf("a disabled signature rendered %q", got)
	}
	if got := signature.EmailText(signature.EmailProps{}); got != "" {
		t.Fatalf("a disabled signature wrote %q", got)
	}
}

// An image of an e-mail is fetched from outside the application, so an origin
// that is not an absolute https address draws nothing instead of a broken
// image.
func TestTheEmailLineNeedsAnAbsoluteHTTPSOrigin(t *testing.T) {
	t.Parallel()

	for _, origin := range []string{"", "letters.example.test", "http://letters.example.test", "https://letters.example.test/path", "https://"} {
		if got := signature.Email(signature.EmailProps{Enabled: true, Origin: origin}); got != "" {
			t.Errorf("origin %q rendered %q", origin, got)
		}
	}
	if got := signature.Email(signature.EmailProps{Enabled: true, Origin: letterOrigin + "/"}); got == "" {
		t.Error("an origin with a trailing slash was refused")
	}
}

// The line reads as the family sentence in one row of a table, each mark an
// image of a fixed size with its name as alternative text, inside a link to its
// site, fetched from the public site: two images for the light theme and two for
// the dark one, and never inline data.
func TestTheEmailLineIsTheFamilySignatureInATable(t *testing.T) {
	t.Parallel()

	html := string(signature.Email(signature.EmailProps{Enabled: true, Origin: letterOrigin, Font: "Montserrat,Arial,sans-serif", Size: 12}))
	for _, want := range []string{
		`<table role="presentation" align="center"`,
		`data-arandu-signature-mail`,
		`lang="pt-BR"`,
		`>Desenvolvido por</td>`,
		`>com</td>`,
		`href="https://hyz.is" title="HYZIS"`,
		`href="https://arandu.io" title="Arandu"`,
		`alt="HYZIS"`,
		`alt="Arandu"`,
		`width="19" height="22"`,
		`width="22" height="20"`,
		`font-family:Montserrat,Arial,sans-serif`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("the line lacks %s:\n%s", want, html)
		}
	}

	images := regexp.MustCompile(`<img\b[^>]*>`).FindAllString(html, -1)
	if len(images) != 4 {
		t.Fatalf("the line has %d images, want the two marks in each of the two themes:\n%s", len(images), html)
	}
	for _, img := range images {
		if !strings.Contains(img, ` alt="`) {
			t.Errorf("an image without alt: %s", img)
		}
		if !regexp.MustCompile(`src="` + regexp.QuoteMeta(letterOrigin) + `/[^"]+/signature-(hyzis|arandu)-mail-(light|dark)\.png"`).MatchString(img) {
			t.Errorf("an image that is not served by the site: %s", img)
		}
		if !regexp.MustCompile(`width="\d+" height="\d+"`).MatchString(img) {
			t.Errorf("an image with no fixed size: %s", img)
		}
	}
	if strings.Count(html, `class="signature-mark signature-dark"`) != 2 || strings.Count(html, `class="signature-mark signature-light"`) != 2 {
		t.Errorf("each mark has a light and a dark image:\n%s", html)
	}
	for _, refused := range []string{"<script", "<link", "data:image", "@import", "javascript:"} {
		if strings.Contains(html, refused) {
			t.Errorf("the line carries %s", refused)
		}
	}
	// A double quote inside a style attribute ends it and drops the rest.
	if regexp.MustCompile(`style="[^"]*"[a-zA-Z ]`).MatchString(html) {
		t.Errorf("a style attribute ended early:\n%s", html)
	}
}

func TestTheEmailLineSpeaksEnglishAlignsLeftAndLinksElsewhere(t *testing.T) {
	t.Parallel()

	html := string(signature.Email(signature.EmailProps{Enabled: true, Origin: letterOrigin, Lang: signature.English, Left: true, HyzisURL: "https://example.test/h"}))
	for _, want := range []string{`>Built by</td>`, `>with</td>`, `lang="en"`, `href="https://example.test/h"`, `margin:0;`} {
		if !strings.Contains(html, want) {
			t.Errorf("the line lacks %s:\n%s", want, html)
		}
	}
	if strings.Contains(html, `align="center"`) {
		t.Errorf("a left line is centred:\n%s", html)
	}
	if got := signature.Email(signature.EmailProps{Enabled: true, Origin: letterOrigin, HyzisURL: "http://example.test"}); got != "" {
		t.Errorf("a link that is not https rendered %q", got)
	}
}

func TestTheEmailTextNamesBothSites(t *testing.T) {
	t.Parallel()

	if got, want := signature.EmailText(signature.EmailProps{Enabled: true}), "Desenvolvido por HYZIS (https://hyz.is) com Arandu (https://arandu.io)"; got != want {
		t.Errorf("text = %q, want %q", got, want)
	}
	if got, want := signature.EmailText(signature.EmailProps{Enabled: true, Lang: signature.English}), "Built by HYZIS (https://hyz.is) with Arandu (https://arandu.io)"; got != want {
		t.Errorf("text = %q, want %q", got, want)
	}
}

// The stylesheet that goes with the line swaps the marks in the dark theme, in
// the rules each client understands, and stacks the four cells on a narrow
// screen.
func TestTheEmailStyleCarriesTheDarkThemeAndTheNarrowScreen(t *testing.T) {
	t.Parallel()

	css := signature.EmailStyle()
	for _, want := range []string{
		`@media (prefers-color-scheme:dark)`,
		`.signature-light{display:none !important`,
		`.signature-dark{display:block !important`,
		`[data-ogsc] .signature-text{color:` + signature.EmailInkDark,
		`@media (max-width:620px)`,
		`td.signature-cell{display:block !important`,
		`text-align:center !important`,
	} {
		if !strings.Contains(css, want) {
			t.Errorf("the style lacks %s:\n%s", want, css)
		}
	}
	if strings.Contains(css, "%!") {
		t.Errorf("the style was formatted wrongly:\n%s", css)
	}
}

// Every image the line names is a view asset the framework serves, as a PNG of
// three times the size it is drawn at, with a transparent background, in the
// ink of the words beside it.
func TestTheEmailMarksAreServedAsPNGInTheInkOfTheWords(t *testing.T) {
	t.Parallel()

	marks := []struct {
		asset, ink string
		w, h       int
	}{
		{signature.HyzisMailLightAsset, signature.EmailInk, signature.EmailHyzisWidth, signature.EmailHyzisHeight},
		{signature.HyzisMailDarkAsset, signature.EmailInkDark, signature.EmailHyzisWidth, signature.EmailHyzisHeight},
		{signature.AranduMailLightAsset, signature.EmailInk, signature.EmailAranduWidth, signature.EmailAranduHeight},
		{signature.AranduMailDarkAsset, signature.EmailInkDark, signature.EmailAranduWidth, signature.EmailAranduHeight},
	}
	for _, m := range marks {
		rec := httptest.NewRecorder()
		view.Handler(rec, httptest.NewRequest(http.MethodGet, view.Asset(m.asset), nil))
		if rec.Code != http.StatusOK || rec.Header().Get("Content-Type") != "image/png" {
			t.Errorf("%s answered %d %q", m.asset, rec.Code, rec.Header().Get("Content-Type"))
			continue
		}
		img, err := png.Decode(bytes.NewReader(rec.Body.Bytes()))
		if err != nil {
			t.Errorf("%s is not a PNG: %v", m.asset, err)
			continue
		}
		if b := img.Bounds(); b.Dx() != 3*m.w || b.Dy() != 3*m.h {
			t.Errorf("%s is %dx%d, want %dx%d (three times the size it is drawn at)", m.asset, b.Dx(), b.Dy(), 3*m.w, 3*m.h)
		}
		want := inkHexRGB(t, m.ink)
		opaque, transparent := 0, 0
		b := img.Bounds()
		for y := b.Min.Y; y < b.Max.Y; y++ {
			for x := b.Min.X; x < b.Max.X; x++ {
				r, g, bl, a := img.At(x, y).RGBA()
				if a == 0 {
					transparent++
					continue
				}
				if a == 0xffff {
					opaque++
					if [3]uint8{uint8(r >> 8), uint8(g >> 8), uint8(bl >> 8)} != want {
						t.Fatalf("%s: an opaque pixel at %d,%d is not %s", m.asset, x, y, m.ink)
					}
				}
			}
		}
		if opaque == 0 || transparent == 0 {
			t.Errorf("%s has %d opaque and %d transparent pixels: the mark needs both", m.asset, opaque, transparent)
		}
	}
}

// The words and the marks read at 4.5:1 or better on the page and the card of
// both themes of a letter (the neutrals of the site).
func TestTheEmailInkReadsOnBothThemes(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct{ ink, background string }{
		{signature.EmailInk, "#ffffff"}, {signature.EmailInk, "#f6f6f6"},
		{signature.EmailInkDark, "#0a0a0a"}, {signature.EmailInkDark, "#171717"},
	} {
		a, b := inkLuminance(t, tc.ink), inkLuminance(t, tc.background)
		if a < b {
			a, b = b, a
		}
		if ratio := (a + 0.05) / (b + 0.05); ratio < 4.5 {
			t.Errorf("%s on %s is %.2f:1", tc.ink, tc.background, ratio)
		}
	}
}

func inkHexRGB(t *testing.T, hex string) [3]uint8 {
	t.Helper()
	var c [3]uint8
	for i := range c {
		v, err := strconv.ParseUint(hex[1+2*i:3+2*i], 16, 8)
		if err != nil {
			t.Fatal(err)
		}
		c[i] = uint8(v)
	}
	return c
}

func inkLuminance(t *testing.T, hex string) float64 {
	t.Helper()
	var l float64
	for i, v := range inkHexRGB(t, hex) {
		f := float64(v) / 255
		if f <= 0.03928 {
			f /= 12.92
		} else {
			f = math.Pow((f+0.055)/1.055, 2.4)
		}
		l += [3]float64{0.2126, 0.7152, 0.0722}[i] * f
	}
	return l
}
