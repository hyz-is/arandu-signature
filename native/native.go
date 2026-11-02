// Package native is the signature for an Ayra app: the line "Desenvolvido por
// [HYZIS] em [Arandu]" with each mark a link to its site, and the About screen
// it belongs on.
//
// It stands apart from the web half so that a web-only project links no native
// code, and so that an app links no web framework: nothing here imports the
// root package.
//
// The line is one widget, [LineProps], that goes on any screen: the About
// screen ("Sobre", reached from the account screen), which carries the app's
// mark, its name and version, one sentence about it and, at its foot, the
// signature; and, where the product is a showcase of HYZIS, the foot of the
// sign-in screen. It never goes next to signing out.
package native

import (
	"bytes"
	_ "embed"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"strings"
	"sync"
	"time"

	"github.com/arandu-io/ayra"
	"github.com/arandu-io/ayra/engine/f32"
	"github.com/arandu-io/ayra/engine/layout"
	"github.com/arandu-io/ayra/engine/op"
	"github.com/arandu-io/ayra/engine/op/paint"
	"github.com/arandu-io/ayra/engine/unit"
	"github.com/arandu-io/ayra/widget"

	"github.com/hyz-is/arandu-signature/internal/licence"
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

	// License is the partnership licence HYZIS issued for the project, kept
	// as a constant in the app's source. A valid one draws the line always:
	// Enabled is ignored. An invalid or expired one is ignored.
	License string
	// Host is the server the app talks to, whose host the licence has to
	// name. Empty checks the signature and the expiry only.
	Host string

	// Height is the HYZIS mark's height; the Arandu mark keeps its
	// proportion to it, 17 to 20. Zero is 20 dp. The target around each mark
	// stays at least 40 dp square.
	Height unit.Dp
}

// heights answers the two marks' heights.
func (o Options) heights() (hyzis, arandu unit.Dp) {
	h := o.Height
	if h <= 0 {
		h = 20
	}
	return h, h * 17 / 20
}

// signs reports whether the line is drawn: a valid licence, or Enabled.
func (o Options) signs() bool {
	if o.License != "" {
		if _, err := licence.Verify(o.License, publicKey, hostOf(o.Host), time.Now()); err == nil {
			return true
		}
	}
	return o.Enabled
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
// The screen holds one per line it draws; one value drawn twice is one line in
// two places.
type State struct {
	hyzis, arandu widget.Clickable
}

// LineProps is how a screen asks for the signature: the same fields as
// [Options], under the name the other widgets of an Ayra app use, so that the
// line is placed as they are.
type LineProps = Options

// markPad is the room around a figure, on every side, that is part of the
// layout. The press area is grown to [widget.TouchTarget] around it.
const markPad unit.Dp = 4

// open is what a tap on a mark calls: [OpenURL] off the drawing goroutine,
// which a test replaces with a recorder.
var open = func(url string) { go func() { _ = OpenURL(url) }() }

// Layout draws the signature centred across the room it is given, in the ink
// of the screen around it, and opens a mark's site when it is tapped. It draws
// nothing, and takes no room, when the project does not sign. Place it where
// the screen puts a footer:
//
//	var signed sig.State
//
//	func layoutLogin(c ayra.Context) ayra.Dimensions {
//		return layout.Flex{Axis: layout.Vertical}.Layout(c.Context,
//			layout.Flexed(1, form),
//			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
//				return sig.LineProps{Enabled: true}.Layout(c.With(gtx), &signed)
//			}),
//		)
//	}
//
// The room reported is the line's own, as high as its figures and the padding
// around them; the area each mark answers to is 44 dp each way, grown outward
// without moving the layout.
func (o Options) Layout(c ayra.Context, s *State) ayra.Dimensions {
	if !o.signs() {
		return ayra.Dimensions{}
	}
	gtx := c.Context
	if gtx.Constraints.Max.X < 1<<24 {
		gtx.Constraints.Min.X = gtx.Constraints.Max.X
	}
	gtx.Constraints.Min.Y = 0
	return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return Line(c.With(gtx), s, o)
	})
}

// Line draws the signature at its own width and opens a mark's site when it is
// tapped. Each mark answers a press within 44 dp. [Options.Layout] is the same
// line centred across the room it is given.
func Line(c ayra.Context, s *State, o Options) ayra.Dimensions {
	if !o.signs() {
		return ayra.Dimensions{}
	}
	hyzisURL, aranduURL := o.urls()
	if s.hyzis.Clicked(c) {
		open(hyzisURL)
	}
	if s.arandu.Clicked(c) {
		open(aranduURL)
	}

	by, with := o.words()
	hyzisHeight, aranduHeight := o.heights()
	ink := c.Theme.Colours.Foreground
	ink.A = 0xBF
	words := func(text string) layout.FlexChild {
		return layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return widget.TextProps{Content: text, Role: widget.Caption, Tone: widget.Muted, MaxLines: 1}.Layout(c.With(gtx))
		})
	}
	mark := func(click *widget.Clickable, name string, src []byte, height unit.Dp) layout.FlexChild {
		return layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return click.Reach(c.With(gtx), func(c ayra.Context) ayra.Dimensions {
				pad := c.Dp(markPad)
				size := markSize(src, c.Dp(height))
				// An even room: the touch area grows it by half of what it lacks
				// on each side, and an odd room would come out a pixel short.
				room := image.Pt(even(size.X+2*pad), even(size.Y+2*pad))
				defer widget.Announce(c, room, widget.Announcement{Label: name, Class: widget.Pressable})()
				defer op.Offset(room.Sub(size).Div(2)).Push(c.Ops).Pop()
				drawMark(c.Context, src, ink, size.Y)
				return ayra.Dimensions{Size: room}
			})
		})
	}
	return layout.Flex{Alignment: layout.Middle}.Layout(c.Context,
		words(by),
		mark(&s.hyzis, "HYZIS", hyzisPNG, hyzisHeight),
		words(with),
		mark(&s.arandu, "Arandu", aranduPNG, aranduHeight),
	)
}

// AboutProps is the About screen of an app.
type AboutProps struct {
	// Back is the app's own way back, drawn first: its "Voltar" button or top
	// bar. Nil draws none, for an app whose navigation draws it elsewhere.
	Back ayra.Widget
	// Eyebrow is the small line above the content, as the app's other screens
	// title themselves ("Sobre o app"). Empty draws none.
	Eyebrow string
	// Mark draws the app's own mark and wordmark, which only the app knows.
	Mark ayra.Widget
	// Name, Version and Description are the app's: "Nivela", "0.1.2",
	// one sentence about what it is for. Name is optional: leave it empty
	// when Mark already carries the wordmark, so the name is not said twice.
	Name, Version, Description string
	// Signature is how the screen signs.
	Signature Options
	// Placement is where the signature goes: AtFoot, the default, pushes it
	// to the bottom of the screen; AfterContent puts it right after the
	// description, centred.
	Placement Placement
}

// Placement is where the About screen puts the signature.
type Placement int

// The places the signature can take on the About screen.
const (
	// AtFoot is the bottom of the screen.
	AtFoot Placement = iota
	// AfterContent is right after the description, centred.
	AfterContent
)

// About draws the About screen: the app's way back and eyebrow when it hands
// them, its mark, name and version, the sentence about it, and the signature
// at the foot of the screen or right after the content.
func About(c ayra.Context, s *State, p AboutProps) ayra.Dimensions {
	gap := func(dp unit.Dp) layout.FlexChild {
		return layout.Rigid(layout.Spacer{Height: dp}.Layout)
	}
	text := func(props widget.TextProps) layout.FlexChild {
		return layout.Rigid(func(gtx layout.Context) layout.Dimensions { return props.Layout(c.With(gtx)) })
	}
	children := []layout.FlexChild{}
	if p.Back != nil {
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions { return p.Back(c.With(gtx)) }), gap(8))
	}
	if p.Eyebrow != "" {
		children = append(children, text(widget.TextProps{Content: strings.ToUpper(p.Eyebrow), Role: widget.Caption, Tone: widget.Muted, MaxLines: 1}), gap(12))
	}
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
	signature := layout.Rigid(func(gtx layout.Context) layout.Dimensions {
		return p.Signature.Layout(c.With(gtx), s)
	})
	if p.Placement == AfterContent {
		children = append(children, gap(32), signature)
	} else {
		children = append(children,
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				return layout.Dimensions{Size: image.Pt(gtx.Constraints.Min.X, gtx.Constraints.Min.Y)}
			}),
			signature,
		)
	}
	return layout.Flex{Axis: layout.Vertical}.Layout(c.Context, children...)
}

// even rounds up to an even number of pixels.
func even(n int) int { return n + n%2 }

// markSize is the size a mark is drawn at for a height in pixels.
func markSize(src []byte, height int) image.Point {
	alpha := alphaOf(src)
	if alpha == nil || height <= 0 {
		return image.Point{}
	}
	b := alpha.Bounds()
	return image.Pt(int(float32(b.Dx())*float32(height)/float32(b.Dy())+0.5), height)
}

// drawMark paints a mark's alpha in ink at the given height, keeping its
// proportions. A mark drawn smaller than its source is resampled to the exact
// pixels it covers and laid on them one to one: scaled by the compositor
// instead, a 96 px figure at 20 dp on a plain screen drops most of its rows
// and the edges come out ragged.
func drawMark(gtx layout.Context, src []byte, ink color.NRGBA, height int) layout.Dimensions {
	img := tinted(src, ink, height)
	if img == nil || height <= 0 {
		return layout.Dimensions{}
	}
	b := img.Bounds()
	if b.Dy() == height {
		paint.NewImageOp(img).Add(gtx.Ops)
		paint.PaintOp{}.Add(gtx.Ops)
		return layout.Dimensions{Size: b.Size()}
	}
	scale := float32(height) / float32(b.Dy())
	width := int(float32(b.Dx())*scale + 0.5)
	defer op.Affine(f32.Affine2D{}.Scale(f32.Point{}, f32.Pt(scale, scale))).Push(gtx.Ops).Pop()
	paint.NewImageOp(img).Add(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)
	return layout.Dimensions{Size: image.Pt(width, height)}
}

// tinted is a mark painted in one colour: at the pixel height asked when that
// is smaller than the source, at the source's own size otherwise. It is made
// once per mark, colour and height.
var (
	tintMu sync.Mutex
	tints  = map[tintKey]*image.NRGBA{}
)

type tintKey struct {
	src    *byte
	ink    color.NRGBA
	height int
}

func tinted(src []byte, ink color.NRGBA, height int) *image.NRGBA {
	alpha := alphaOf(src)
	if alpha == nil {
		return nil
	}
	if height >= alpha.Bounds().Dy() {
		height = 0
	}
	key := tintKey{&src[0], ink, height}
	tintMu.Lock()
	defer tintMu.Unlock()
	if img, ok := tints[key]; ok {
		return img
	}
	if height > 0 {
		alpha = shrink(alpha, height)
	}
	b := alpha.Bounds()
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

// alphaOf is the alpha plane of an embedded mark, decoded once.
func alphaOf(src []byte) *image.Alpha {
	alphaMu.Lock()
	defer alphaMu.Unlock()
	if alpha, ok := alphas[&src[0]]; ok {
		return alpha
	}
	decoded, err := png.Decode(bytes.NewReader(src))
	if err != nil {
		alphas[&src[0]] = nil
		return nil
	}
	b := decoded.Bounds()
	alpha := image.NewAlpha(b)
	draw.Draw(alpha, b, decoded, b.Min, draw.Src)
	alphas[&src[0]] = alpha
	return alpha
}

var (
	alphaMu sync.Mutex
	alphas  = map[*byte]*image.Alpha{}
)

// shrink resamples an alpha plane to a smaller height by the area each new
// pixel covers, which keeps a thin stroke as faint as it is thin instead of
// dropping it.
func shrink(src *image.Alpha, height int) *image.Alpha {
	b := src.Bounds()
	width := int(float64(b.Dx())*float64(height)/float64(b.Dy()) + 0.5)
	if width < 1 {
		width = 1
	}
	out := image.NewAlpha(image.Rect(0, 0, width, height))
	sx, sy := float64(b.Dx())/float64(width), float64(b.Dy())/float64(height)
	for y := 0; y < height; y++ {
		y0, y1 := float64(y)*sy, float64(y+1)*sy
		for x := 0; x < width; x++ {
			x0, x1 := float64(x)*sx, float64(x+1)*sx
			var sum, weight float64
			for j := int(y0); float64(j) < y1 && j < b.Dy(); j++ {
				wy := min(y1, float64(j+1)) - max(y0, float64(j))
				for i := int(x0); float64(i) < x1 && i < b.Dx(); i++ {
					w := wy * (min(x1, float64(i+1)) - max(x0, float64(i)))
					sum += w * float64(src.AlphaAt(b.Min.X+i, b.Min.Y+j).A)
					weight += w
				}
			}
			if weight > 0 {
				out.SetAlpha(x, y, color.Alpha{A: uint8(sum/weight + 0.5)})
			}
		}
	}
	return out
}
