// Package native is the signature for an Ayra app: the line "Desenvolvido por
// [HYZIS] em [Arandu]" with each mark a link to its site, and the About screen
// it belongs on.
//
// It stands apart from the web half so that a web-only project links no native
// code, and so that an app links no web framework: nothing here imports the
// root package.
//
// The line goes on an About screen, never next to signing out: the account
// screen offers "Sobre", and About carries the app's mark, its name and
// version, one sentence about it and, at its foot, the signature.
package native

import (
	"bytes"
	_ "embed"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"sync"

	"github.com/arandu-io/ayra"
	"github.com/arandu-io/ayra/engine/f32"
	"github.com/arandu-io/ayra/engine/layout"
	"github.com/arandu-io/ayra/engine/op"
	"github.com/arandu-io/ayra/engine/op/paint"
	"github.com/arandu-io/ayra/engine/unit"
	ewidget "github.com/arandu-io/ayra/engine/widget"
	"github.com/arandu-io/ayra/widget"
)

// The official marks, rasterized at 96 px tall with a transparent background
// from the same files the web half serves. Only their alpha is used: each is
// painted in the ink of the screen around it, as the site masks them.
var (
	//go:embed marks/hyzis.png
	hyzisPNG []byte
	//go:embed marks/arandu.png
	aranduPNG []byte
)

// The two addresses the marks link to.
const (
	DefaultHyzisURL  = "https://hyz.is"
	DefaultAranduURL = "https://arandu.io"
)

// Options is how the app signs.
type Options struct {
	// Enabled is the contract: the line appears only on a product whose client
	// agreed to the brand partnership. Zero draws nothing.
	Enabled bool
	// English reads "Built by … with …". Zero reads "Desenvolvido por … em …".
	English bool
	// HyzisURL and AranduURL override where the marks link to. Empty is the
	// default address of each.
	HyzisURL  string
	AranduURL string
}

func (o Options) words() (by, with string) {
	if o.English {
		return "Built by", "with"
	}
	return "Desenvolvido por", "em"
}

func (o Options) urls() (hyzis, arandu string) {
	hyzis, arandu = o.HyzisURL, o.AranduURL
	if hyzis == "" {
		hyzis = DefaultHyzisURL
	}
	if arandu == "" {
		arandu = DefaultAranduURL
	}
	return hyzis, arandu
}

// State is what the line remembers between frames: the two marks as targets.
type State struct {
	hyzis, arandu ewidget.Clickable
}

// Line draws the signature and opens a mark's site when it is tapped. A mark
// is a target at least 40 dp square around a figure as tall as the text.
func Line(c ayra.Context, s *State, o Options) ayra.Dimensions {
	if !o.Enabled {
		return ayra.Dimensions{}
	}
	hyzisURL, aranduURL := o.urls()
	if s.hyzis.Clicked(c.Context) {
		go func() { _ = OpenURL(hyzisURL) }()
	}
	if s.arandu.Clicked(c.Context) {
		go func() { _ = OpenURL(aranduURL) }()
	}

	by, with := o.words()
	ink := c.Theme.Colours.Foreground
	ink.A = 0xBF
	words := func(text string) layout.FlexChild {
		return layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return widget.TextProps{Content: text, Role: widget.Caption, Tone: widget.Muted, MaxLines: 1}.Layout(c.With(gtx))
		})
	}
	mark := func(click *ewidget.Clickable, src []byte, height unit.Dp) layout.FlexChild {
		return layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return click.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				side := gtx.Dp(40)
				gtx.Constraints.Min = image.Pt(side, side)
				return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return drawMark(gtx, src, ink, gtx.Dp(height))
				})
			})
		})
	}
	return layout.Flex{Alignment: layout.Middle}.Layout(c.Context,
		words(by),
		mark(&s.hyzis, hyzisPNG, 20),
		words(with),
		mark(&s.arandu, aranduPNG, 17),
	)
}

// AboutProps is the About screen of an app.
type AboutProps struct {
	// Mark draws the app's own mark and wordmark, which only the app knows.
	Mark ayra.Widget
	// Name, Version and Description are the app's: "Nivela", "0.1.2",
	// one sentence about what it is for.
	Name, Version, Description string
	// Signature is how the foot of the screen signs.
	Signature Options
}

// About draws the About screen: the app's mark, its name and version, the
// sentence about it, and the signature at the foot of the screen.
func About(c ayra.Context, s *State, p AboutProps) ayra.Dimensions {
	gap := func(dp unit.Dp) layout.FlexChild {
		return layout.Rigid(layout.Spacer{Height: dp}.Layout)
	}
	text := func(props widget.TextProps) layout.FlexChild {
		return layout.Rigid(func(gtx layout.Context) layout.Dimensions { return props.Layout(c.With(gtx)) })
	}
	children := []layout.FlexChild{}
	if p.Mark != nil {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions { return p.Mark(c.With(gtx)) }), gap(16))
	}
	if p.Name != "" {
		children = append(children, text(widget.TextProps{Content: p.Name, Role: widget.Heading}))
	}
	if p.Version != "" {
		version := "Versão " + p.Version
		if p.Signature.English {
			version = "Version " + p.Version
		}
		children = append(children, gap(4), text(widget.TextProps{Content: version, Role: widget.Caption, Tone: widget.Muted}))
	}
	if p.Description != "" {
		children = append(children, gap(12), text(widget.TextProps{Content: p.Description}))
	}
	children = append(children,
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return layout.Dimensions{Size: image.Pt(gtx.Constraints.Min.X, gtx.Constraints.Min.Y)}
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions { return Line(c.With(gtx), s, p.Signature) })
		}),
	)
	return layout.Flex{Axis: layout.Vertical}.Layout(c.Context, children...)
}

// drawMark paints a mark's alpha in ink at the given height, keeping its
// proportions.
func drawMark(gtx layout.Context, src []byte, ink color.NRGBA, height int) layout.Dimensions {
	img := tinted(src, ink)
	if img == nil || height <= 0 {
		return layout.Dimensions{}
	}
	b := img.Bounds()
	scale := float32(height) / float32(b.Dy())
	width := int(float32(b.Dx())*scale + 0.5)
	defer op.Affine(f32.Affine2D{}.Scale(f32.Point{}, f32.Pt(scale, scale))).Push(gtx.Ops).Pop()
	paint.NewImageOp(img).Add(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)
	return layout.Dimensions{Size: image.Pt(width, height)}
}

// tinted is a mark painted in one colour, decoded once per mark and colour.
var (
	tintMu sync.Mutex
	tints  = map[tintKey]*image.NRGBA{}
)

type tintKey struct {
	src *byte
	ink color.NRGBA
}

func tinted(src []byte, ink color.NRGBA) *image.NRGBA {
	key := tintKey{&src[0], ink}
	tintMu.Lock()
	defer tintMu.Unlock()
	if img, ok := tints[key]; ok {
		return img
	}
	decoded, err := png.Decode(bytes.NewReader(src))
	if err != nil {
		return nil
	}
	b := decoded.Bounds()
	alpha := image.NewAlpha(b)
	draw.Draw(alpha, b, decoded, b.Min, draw.Src)
	out := image.NewNRGBA(b)
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			a := alpha.AlphaAt(x, y).A
			out.SetNRGBA(x, y, color.NRGBA{R: ink.R, G: ink.G, B: ink.B, A: uint8(uint16(a) * uint16(ink.A) / 255)})
		}
	}
	tints[key] = out
	return out
}
