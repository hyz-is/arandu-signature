package signature

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"

	"github.com/arandu-io/hesape/view"
)

// The official marks, as the sibling products serve them.
var (
	//go:embed marks/hyzis.svg
	hyzisMark []byte
	//go:embed marks/arandu.svg
	aranduMark []byte
)

// The names the three files are registered under as view assets.
const (
	HyzisAsset      = "signature-hyzis.svg"
	AranduAsset     = "signature-arandu.svg"
	StylesheetAsset = "signature.css"
)

// The stylesheet is written once, after the marks are registered, because the
// masks name the marks by the hashed address the framework serves them at.
func init() {
	view.RegisterAsset(HyzisAsset, "image/svg+xml", hyzisMark)
	view.RegisterAsset(AranduAsset, "image/svg+xml", aranduMark)
	view.RegisterAsset(StylesheetAsset, "text/css; charset=utf-8", stylesheet(view.Asset(HyzisAsset), view.Asset(AranduAsset)))
}

// stylesheet is the line's own style: the marks are masks filled with the ink
// around them, as large as the text, and each link has a target of at least
// 24 px for a finger.
func stylesheet(hyzis, arandu string) []byte {
	return []byte(fmt.Sprintf(`.arandu-signature { display: inline-flex; flex-wrap: wrap; align-items: center; gap: 0.6em; margin: 0; color: inherit; font-size: inherit; }
.arandu-signature a { display: inline-flex; align-items: center; min-height: 24px; min-width: 24px; justify-content: center; color: color-mix(in oklch, currentColor 75%%, transparent); text-decoration: none; }
.arandu-signature a:hover, .arandu-signature a:focus-visible { color: currentColor; }
.arandu-signature-mark { display: block; background-color: currentColor; -webkit-mask-repeat: no-repeat; mask-repeat: no-repeat; -webkit-mask-position: center; mask-position: center; -webkit-mask-size: contain; mask-size: contain; }
.arandu-signature-hyzis { height: 1.4em; aspect-ratio: 1876 / 2168; -webkit-mask-image: url("%[1]s"); mask-image: url("%[1]s"); }
.arandu-signature-arandu { height: 1.15em; aspect-ratio: 292 / 260; -webkit-mask-image: url("%[2]s"); mask-image: url("%[2]s"); }
`, hyzis, arandu))
}

// Props is one rendering of the line.
type Props struct {
	// Enabled is the contract. Zero renders nothing.
	Enabled bool
	// Lang is PortugueseBR or English. Empty is PortugueseBR.
	Lang string
	// HyzisURL and AranduURL override where the marks link to. Empty is the
	// default address of each.
	HyzisURL  string
	AranduURL string
	// Class is added to the line's element, for the layout's own placement.
	Class string
}

var line = template.Must(template.New("signature").Parse(
	`<link rel="stylesheet" href="{{.Stylesheet}}">` +
		`<p class="arandu-signature{{if .Class}} {{.Class}}{{end}}" lang="{{.Lang}}">` +
		`<span>{{.By}}</span>` +
		`<a href="{{.HyzisURL}}" aria-label="HYZIS" rel="noopener"><span class="arandu-signature-mark arandu-signature-hyzis" aria-hidden="true"></span></a>` +
		`<span>{{.With}}</span>` +
		`<a href="{{.AranduURL}}" aria-label="Arandu" rel="noopener"><span class="arandu-signature-mark arandu-signature-arandu" aria-hidden="true"></span></a>` +
		`</p>`))

// Kyse is the line as markup for a Kyse layout:
//
//	{!! signature.Kyse(signature.Props{Enabled: true}) !!}
//
// It carries its own stylesheet, served from the origin, so the layout adds
// nothing else and a strict Content-Security-Policy holds: no inline style and
// no script. A project that does not sign gets nothing.
func Kyse(p Props) template.HTML {
	if !p.Enabled {
		return ""
	}
	cfg := Config{Enabled: true, Lang: p.Lang, HyzisURL: p.HyzisURL, AranduURL: p.AranduURL}
	if cfg.Validate() != nil {
		return ""
	}
	cfg = cfg.withDefaults()
	by, with := Words(cfg.Lang)
	var b bytes.Buffer
	_ = line.Execute(&b, struct {
		Stylesheet, Class, Lang, By, With, HyzisURL, AranduURL string
	}{view.Asset(StylesheetAsset), p.Class, cfg.Lang, by, with, cfg.HyzisURL, cfg.AranduURL})
	return template.HTML(b.String())
}
