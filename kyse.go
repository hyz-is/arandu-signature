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

// Marker is the attribute every rendering of the line carries. Middleware
// looks for it to know whether a page still signs.
const Marker = "data-arandu-signature"

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
//
// The rules sit on the element and its attribute, and every property that
// could make the line disappear is !important: a page's own rule that hides
// paragraphs, shrinks text, moves the footer off screen or fades it out does
// not win against them. It is friction, not a lock -- whoever owns the source
// can still delete the line, and that is a change to the source.
func stylesheet(hyzis, arandu string) []byte {
	return []byte(fmt.Sprintf(`p.arandu-signature[data-arandu-signature], p.arandu-signature[data-arandu-signature] * { visibility: visible !important; opacity: 1 !important; filter: none !important; transform: none !important; clip: auto !important; clip-path: none !important; mask-clip: border-box; }
p.arandu-signature[data-arandu-signature] { display: inline-flex !important; flex-wrap: wrap !important; align-items: center !important; gap: 0.6em !important; position: static !important; width: auto !important; height: auto !important; min-width: 0; max-width: none !important; max-height: none !important; overflow: visible !important; margin: 0; padding: 0; color: inherit; font-size: max(12px, 1em) !important; line-height: 1.4 !important; text-indent: 0 !important; z-index: auto !important; pointer-events: auto !important; user-select: auto; }
p.arandu-signature[data-arandu-signature] > span:not(.arandu-signature-mark) { display: inline !important; font-size: inherit !important; color: inherit; }
p.arandu-signature[data-arandu-signature] a { display: inline-flex !important; align-items: center; min-height: 24px !important; min-width: 24px !important; justify-content: center; color: color-mix(in oklch, currentColor 75%%, transparent); text-decoration: none; pointer-events: auto !important; }
p.arandu-signature[data-arandu-signature] a:hover, p.arandu-signature[data-arandu-signature] a:focus-visible { color: currentColor; }
p.arandu-signature[data-arandu-signature] .arandu-signature-mark { display: block !important; background-color: currentColor !important; -webkit-mask-repeat: no-repeat; mask-repeat: no-repeat; -webkit-mask-position: center; mask-position: center; -webkit-mask-size: contain; mask-size: contain; }
p.arandu-signature[data-arandu-signature] .arandu-signature-hyzis { height: 1.4em !important; min-height: 16px !important; aspect-ratio: 1876 / 2168; -webkit-mask-image: url("%[1]s") !important; mask-image: url("%[1]s") !important; }
p.arandu-signature[data-arandu-signature] .arandu-signature-arandu { height: 1.15em !important; min-height: 13px !important; aspect-ratio: 292 / 260; -webkit-mask-image: url("%[2]s") !important; mask-image: url("%[2]s") !important; }
div.arandu-signature-restored[data-arandu-signature-restored] { display: flex !important; justify-content: center !important; padding: 1.25rem 1rem !important; visibility: visible !important; opacity: 1 !important; position: static !important; }
`, hyzis, arandu))
}

// Props is one rendering of the line.
type Props struct {
	// Enabled is the contract. Zero renders nothing, unless the project holds
	// a partnership licence: then the line is always drawn.
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
		`<p class="arandu-signature{{if .Class}} {{.Class}}{{end}}" data-arandu-signature lang="{{.Lang}}">` +
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
// no script. A project that does not sign gets nothing; a project with a
// partnership licence always gets the line, with the licensed addresses and
// language when the call names none.
func Kyse(p Props) template.HTML {
	if licensed := forced.Load(); licensed != nil {
		p.Enabled = true
		if p.Lang == "" {
			p.Lang = licensed.Lang
		}
		if p.HyzisURL == "" {
			p.HyzisURL = licensed.HyzisURL
		}
		if p.AranduURL == "" {
			p.AranduURL = licensed.AranduURL
		}
	}
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
