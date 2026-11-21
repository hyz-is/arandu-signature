package signature

import (
	_ "embed"
	"fmt"
	"html/template"
	"net/url"
	"strings"

	"github.com/arandu-io/hesape/view"
)

// The marks as an e-mail draws them. A mail client draws no SVG and no CSS
// mask, so each mark is a PNG in one flat ink: the muted ink of a letter's
// foot on a light page, and its light counterpart on a dark one. They are the
// same two marks as the web line, rendered from marks/hyzis.svg and
// marks/arandu.svg at three times the size they are drawn at.
var (
	//go:embed marks/hyzis-mail-light.png
	hyzisMailLight []byte
	//go:embed marks/hyzis-mail-dark.png
	hyzisMailDark []byte
	//go:embed marks/arandu-mail-light.png
	aranduMailLight []byte
	//go:embed marks/arandu-mail-dark.png
	aranduMailDark []byte
)

// The names the four PNGs are registered under as view assets. A letter's
// images are fetched from the public site, at Origin plus view.Asset(name).
const (
	HyzisMailLightAsset  = "signature-hyzis-mail-light.png"
	HyzisMailDarkAsset   = "signature-hyzis-mail-dark.png"
	AranduMailLightAsset = "signature-arandu-mail-light.png"
	AranduMailDarkAsset  = "signature-arandu-mail-dark.png"
)

// The size each mark is drawn at in an e-mail, in CSS pixels. The PNGs are
// three times that, and the width and height attributes fix the size because a
// mail client has no stylesheet to be trusted with it.
const (
	EmailHyzisWidth, EmailHyzisHeight   = 19, 22
	EmailAranduWidth, EmailAranduHeight = 22, 20
)

// The ink of the words beside the marks, and of the marks themselves: the
// muted text of a letter's foot, in the light and in the dark theme. Both pass
// 4.5:1 on the page and on the card of either theme.
const (
	EmailInk     = "#525252"
	EmailInkDark = "#a3a3a3"
)

// EmailMarker is the attribute the e-mail form of the line carries.
const EmailMarker = "data-arandu-signature-mail"

func init() {
	view.RegisterAsset(HyzisMailLightAsset, "image/png", hyzisMailLight)
	view.RegisterAsset(HyzisMailDarkAsset, "image/png", hyzisMailDark)
	view.RegisterAsset(AranduMailLightAsset, "image/png", aranduMailLight)
	view.RegisterAsset(AranduMailDarkAsset, "image/png", aranduMailDark)
}

// EmailProps is one rendering of the line for an e-mail.
type EmailProps struct {
	// Enabled is the contract, as in Props: zero renders nothing, unless the
	// project holds a partnership licence.
	Enabled bool
	// Lang is the page's language tag: see [Language]. Empty is PortugueseBR
	// and a language the line has no words for is English. Arabic draws the
	// table right to left.
	Lang string
	// HyzisURL and AranduURL override where the marks link to; each has to be
	// an absolute https address.
	HyzisURL  string
	AranduURL string
	// Origin is the public address the images are fetched from, the origin of
	// the site that serves the view assets ("https://example.com", no path):
	// an e-mail is read outside the application, so the address of an image is
	// absolute. An Origin that is not an absolute https address renders
	// nothing.
	Origin string
	// Font is the font-family of the words, written with single quotes because
	// it sits inside a style attribute. Empty is a system sans-serif stack.
	Font string
	// Size is the size of the words in CSS pixels. Zero is 12.
	Size int
	// Left aligns the line to the left, as the text above it. Empty is the
	// centre, as the line sits on a site.
	Left bool
}

const defaultEmailFont = `-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,Helvetica,Arial,sans-serif`

// resolve applies the licence to a request for the line and answers the
// configuration to draw, or false when there is nothing to draw.
func resolve(enabled bool, lang, hyzisURL, aranduURL string) (Config, bool) {
	if licensed := forced.Load(); licensed != nil {
		enabled = true
		if lang == "" {
			lang = licensed.Lang
		}
		if hyzisURL == "" {
			hyzisURL = licensed.HyzisURL
		}
		if aranduURL == "" {
			aranduURL = licensed.AranduURL
		}
	}
	if !enabled {
		return Config{}, false
	}
	cfg := Config{Enabled: true, Lang: Language(lang), HyzisURL: hyzisURL, AranduURL: aranduURL}
	if cfg.Validate() != nil {
		return Config{}, false
	}
	return cfg.withDefaults(), true
}

// emailOrigin is the origin without a trailing slash, or false when it is not
// an absolute https address.
func emailOrigin(origin string) (string, bool) {
	u, err := url.Parse(origin)
	if err != nil || u.Scheme != "https" || u.Host == "" || (u.Path != "" && u.Path != "/") || u.RawQuery != "" || u.Fragment != "" {
		return "", false
	}
	return "https://" + u.Host, true
}

// Email is the line as markup for an e-mail layout, in the inline styles an
// inbox keeps:
//
//	{!! signature.Email(signature.EmailProps{Enabled: true, Origin: "https://example.com"}) !!}
//
// It is a table: "Desenvolvido por", the HYZIS mark, "com" and the Arandu mark
// in one row (a language with a closing word, Japanese or Chinese, adds a
// last cell for it), each mark a link with its name as alternative text, drawn at a
// fixed size from a PNG served by the site. A letter's own stylesheet adds the
// rest by including EmailStyle: the dark theme swaps each mark for its light
// counterpart, and on a narrow screen the four cells stack, centred, one under
// the other. A client that drops the stylesheet keeps the one row of the light
// theme, which is complete on its own.
//
// A project that does not sign gets nothing; a project with a partnership
// licence always gets the line.
func Email(p EmailProps) template.HTML {
	cfg, ok := resolve(p.Enabled, p.Lang, p.HyzisURL, p.AranduURL)
	if !ok {
		return ""
	}
	origin, ok := emailOrigin(p.Origin)
	if !ok {
		return ""
	}
	font := p.Font
	if font == "" {
		font = defaultEmailFont
	}
	size := p.Size
	if size <= 0 {
		size = 12
	}
	phrase := Sentence(cfg.Lang)
	esc := template.HTMLEscapeString

	align, margin := ` align="center"`, "0 auto"
	if p.Left {
		align, margin = "", "0"
	}
	word := func(text string) string {
		return fmt.Sprintf(`<td class="signature-cell signature-text" valign="middle" style="padding:0 8px;font-family:%s;font-size:%dpx;line-height:1.4;color:%s;white-space:nowrap">%s</td>`,
			font, size, EmailInk, esc(text))
	}
	mark := func(name, href, light, dark string, w, h int) string {
		img := func(class, asset, hidden string) string {
			return fmt.Sprintf(`<img class="signature-mark %s" src="%s" width="%d" height="%d" alt="%s" style="%sborder:0;width:%dpx;height:%dpx;margin:0;color:%s;font-family:%s;font-size:%dpx;line-height:%dpx">`,
				class, esc(origin+view.Asset(asset)), w, h, esc(name), hidden, w, h, EmailInk, font, size, h)
		}
		return fmt.Sprintf(`<td class="signature-cell" valign="middle" style="padding:0 8px"><a href="%s" title="%s" rel="noopener" style="display:block;text-decoration:none">%s%s</a></td>`,
			esc(href), esc(name), img("signature-light", light, "display:block;"), img("signature-dark", dark, "display:none;max-height:0;overflow:hidden;"))
	}

	var b strings.Builder
	fmt.Fprintf(&b, `<table role="presentation"%s cellpadding="0" cellspacing="0" border="0" class="signature-mail" %s lang="%s" dir="%s" style="margin:%s;border-collapse:collapse"><tr>`,
		align, EmailMarker, esc(phrase.HTMLLang), phrase.Dir(), margin)
	words := func(text string) {
		if text != "" {
			b.WriteString(word(text))
		}
	}
	words(phrase.Before)
	b.WriteString(mark("HYZIS", cfg.HyzisURL, HyzisMailLightAsset, HyzisMailDarkAsset, EmailHyzisWidth, EmailHyzisHeight))
	words(phrase.Between)
	b.WriteString(mark("Arandu", cfg.AranduURL, AranduMailLightAsset, AranduMailDarkAsset, EmailAranduWidth, EmailAranduHeight))
	words(phrase.After)
	b.WriteString(`</tr></table>`)
	return template.HTML(b.String()) // #nosec G203 -- every value was escaped above
}

// EmailText is the line for the plain-text part of an e-mail, with the address
// of each mark beside its name, or empty when the project does not sign.
func EmailText(p EmailProps) string {
	cfg, ok := resolve(p.Enabled, p.Lang, p.HyzisURL, p.AranduURL)
	if !ok {
		return ""
	}
	phrase := Sentence(cfg.Lang)
	parts := []string{phrase.Before, "HYZIS (" + cfg.HyzisURL + ")", phrase.Between, "Arandu (" + cfg.AranduURL + ")", phrase.After}
	kept := parts[:0]
	for _, part := range parts {
		if part != "" {
			kept = append(kept, part)
		}
	}
	return strings.Join(kept, " ")
}

// EmailStyle is the rules the e-mail line needs in the <style> of a letter's
// head, which a client that supports a stylesheet reads and one that drops it
// ignores. It carries the dark theme, as the rules each client understands --
// the media query of Apple Mail, Outlook for Mac and iOS and Thunderbird, and
// the data-ogsc marker Outlook.com puts on what it darkens -- and the narrow
// screen, where the four cells stack, centred.
//
//	<style>{{ your rules }}{{ signature.EmailStyle() }}</style>
func EmailStyle() string {
	return fmt.Sprintf(`.signature-dark{display:none !important;max-height:0 !important;overflow:hidden !important}
@media (prefers-color-scheme:dark){
 .signature-text{color:%[1]s !important}
 .signature-light{display:none !important;max-height:0 !important;overflow:hidden !important}
 .signature-dark{display:block !important;max-height:none !important;overflow:visible !important}
}
[data-ogsc] .signature-text{color:%[1]s !important}
@media (max-width:620px){
 table.signature-mail,table.signature-mail tbody,table.signature-mail tr{display:block !important;width:100%% !important;margin:0 !important}
 td.signature-cell{display:block !important;width:100%% !important;padding:0 0 10px !important;text-align:center !important;white-space:normal !important}
 .signature-mark{margin:0 auto !important}
}
`, EmailInkDark)
}
