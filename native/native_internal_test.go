package native

import (
	"image"
	"testing"

	"github.com/arandu-io/ayra"
	"github.com/arandu-io/ayra/engine/f32"
	"github.com/arandu-io/ayra/engine/font/gofont"
	"github.com/arandu-io/ayra/engine/io/input"
	"github.com/arandu-io/ayra/engine/io/pointer"
	"github.com/arandu-io/ayra/engine/layout"
	"github.com/arandu-io/ayra/engine/op"
	"github.com/arandu-io/ayra/engine/text"
	"github.com/arandu-io/ayra/engine/unit"
	"github.com/arandu-io/ayra/theme"
)

// The Arandu mark keeps its 17-to-20 proportion to the HYZIS mark at any
// height, and zero is the default 20 dp.
func TestTheMarksKeepTheirProportion(t *testing.T) {
	for _, tc := range []struct {
		height, hyzis, arandu float32
	}{{0, 20, 17}, {20, 20, 17}, {40, 40, 34}} {
		h, a := Options{Height: unitDp(tc.height)}.heights()
		if float32(h) != tc.hyzis || float32(a) != tc.arandu {
			t.Errorf("height %v drew %v and %v, want %v and %v", tc.height, h, a, tc.hyzis, tc.arandu)
		}
	}
}

// The About screen keeps v0.2.0's place for the signature unless asked.
func TestTheSignatureGoesToTheFootByDefault(t *testing.T) {
	if (AboutProps{}).Placement != AtFoot {
		t.Fatal("the zero About screen does not put the signature at its foot")
	}
}

func unitDp(v float32) unit.Dp { return unit.Dp(v) }

// A press on a mark opens its site, and the area that answers is 44 dp each way
// although the figure and its padding are smaller.
func TestAPressOnAMarkOpensItsSiteWithin44Dp(t *testing.T) {
	var opened []string
	defer func(was func(string)) { open = was }(open)
	open = func(url string) { opened = append(opened, url) }

	size := image.Pt(390, 300)
	shaper := text.NewShaper(text.WithCollection(gofont.Collection()))
	drawn := image.Point{}
	frameOf := func(router *input.Router, ops *op.Ops, state *State) {
		ops.Reset()
		gtx := layout.Context{
			Ops:         ops,
			Source:      router.Source(),
			Metric:      unit.Metric{PxPerDp: 1, PxPerSp: 1},
			Constraints: layout.Constraints{Max: size},
		}
		c := ayra.Context{Context: gtx, Theme: theme.New(theme.Light), Shaper: shaper}
		dims := layout.Inset{Top: 100}.Layout(c.Context, func(gtx layout.Context) layout.Dimensions {
			return LineProps{Enabled: true}.Layout(c.With(gtx), state)
		})
		drawn = dims.Size.Sub(image.Pt(0, 100))
		router.Frame(ops)
	}
	tap := func(at image.Point) []string {
		var router input.Router
		var ops op.Ops
		var state State
		opened = nil
		frameOf(&router, &ops, &state)
		frameOf(&router, &ops, &state)
		pos := f32.Pt(float32(at.X), float32(at.Y))
		router.Queue(pointer.Event{Kind: pointer.Press, Source: pointer.Touch, PointerID: 1, Position: pos})
		frameOf(&router, &ops, &state)
		router.Queue(pointer.Event{Kind: pointer.Release, Source: pointer.Touch, PointerID: 1, Position: pos})
		frameOf(&router, &ops, &state)
		frameOf(&router, &ops, &state)
		return append([]string(nil), opened...)
	}

	tap(image.Pt(0, 0))
	if drawn.X != size.X {
		t.Fatalf("the line took %d points across, want the whole %d: it is centred in the room it is given", drawn.X, size.X)
	}
	if drawn.Y >= 44 {
		t.Fatalf("the line is %d points high, which is no test of a smaller figure", drawn.Y)
	}

	// Across the middle of the line: where each site opens.
	middle := 100 + drawn.Y/2
	var hyzisX, aranduX []int
	for x := 0; x < size.X; x += 2 {
		switch got := tap(image.Pt(x, middle)); {
		case len(got) == 1 && got[0] == DefaultHyzisURL:
			hyzisX = append(hyzisX, x)
		case len(got) == 1 && got[0] == DefaultAranduURL:
			aranduX = append(aranduX, x)
		case len(got) > 1:
			t.Fatalf("one press at x=%d opened %v", x, got)
		}
	}
	if len(hyzisX) == 0 || len(aranduX) == 0 {
		t.Fatalf("a press opened hyzis at %v and arandu at %v: each mark must open its site", hyzisX, aranduX)
	}
	if hyzisX[len(hyzisX)-1] >= aranduX[0] {
		t.Errorf("the HYZIS mark (%d to %d) is not before the Arandu mark (%d to %d)", hyzisX[0], hyzisX[len(hyzisX)-1], aranduX[0], aranduX[len(aranduX)-1])
	}

	// Up and down from the middle of each mark.
	for name, xs := range map[string][]int{DefaultHyzisURL: hyzisX, DefaultAranduURL: aranduX} {
		x := xs[len(xs)/2]
		var ys []int
		for y := 100 - 40; y < 100+drawn.Y+40; y++ {
			if got := tap(image.Pt(x, y)); len(got) == 1 && got[0] == name {
				ys = append(ys, y)
			}
		}
		if height := len(ys); height < 43 {
			t.Errorf("%s answers a press over %d points of height, want at least 44", name, height)
		}
		if ys[len(ys)-1]-ys[0]+1 != len(ys) {
			t.Errorf("%s answers over a broken span: %v", name, ys)
		}
	}
	if width := (hyzisX[len(hyzisX)-1] - hyzisX[0]) + 2; width < 43 {
		t.Errorf("the HYZIS mark answers a press over %d points across, want at least 44", width)
	}
	if width := (aranduX[len(aranduX)-1] - aranduX[0]) + 2; width < 43 {
		t.Errorf("the Arandu mark answers a press over %d points across, want at least 44", width)
	}

	// Off the line, nothing opens.
	if got := tap(image.Pt(size.X/2, 5)); len(got) != 0 {
		t.Errorf("a press far above the line opened %v", got)
	}
}

// A project that does not sign draws nothing, takes no room and answers no press.
func TestAnUnsignedLineTakesNoRoom(t *testing.T) {
	var ops op.Ops
	var router input.Router
	gtx := layout.Context{Ops: &ops, Source: router.Source(), Metric: unit.Metric{PxPerDp: 1, PxPerSp: 1},
		Constraints: layout.Constraints{Max: image.Pt(390, 300)}}
	c := ayra.Context{Context: gtx, Theme: theme.New(theme.Light),
		Shaper: text.NewShaper(text.WithCollection(gofont.Collection()))}
	if dims := (LineProps{}).Layout(c, new(State)); dims.Size != (image.Point{}) {
		t.Errorf("an unsigned line took %v", dims.Size)
	}
}
