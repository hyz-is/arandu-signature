package unit_test

import (
	"flag"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/arandu-io/ayra"
	"github.com/arandu-io/ayra/engine/font/gofont"
	"github.com/arandu-io/ayra/engine/layout"
	"github.com/arandu-io/ayra/frame"
	"github.com/arandu-io/ayra/theme"

	sig "github.com/hyz-is/arandu-signature/native"
)

var update = flag.Bool("update", false, "rewrite the approved pictures of the native line")

// footer is a screen with the line where a sign-in screen puts it: at the foot,
// centred across a phone's width.
func footer() ayra.Widget {
	var state sig.State
	return func(c ayra.Context) ayra.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(c.Context,
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				return layout.Dimensions{Size: gtx.Constraints.Max}
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return sig.LineProps{Enabled: true}.Layout(c.With(gtx), &state)
			}),
		)
	}
}

func drawFooter(t *testing.T, scheme theme.Scheme, density float32) *image.RGBA {
	t.Helper()
	got, err := frame.Draw(frame.Options{Width: 390, Height: 96, Density: density, Scheme: scheme, Fonts: gofont.Collection()}, footer())
	if err != nil {
		if strings.Contains(err.Error(), "no drawing surface") {
			t.Skipf("no drawing surface on this machine: %v", err)
		}
		t.Fatalf("the screen did not draw: %v", err)
	}
	return got
}

// The line looks as it was approved, in both schemes, at 390 dp on an ordinary
// screen and on a dense one. `go test ./tests/Unit -run NativeLine -update`
// rewrites the pictures; look at them before committing.
func TestTheNativeLineLooksAsApproved(t *testing.T) {
	for _, scheme := range []theme.Scheme{theme.Light, theme.Dark} {
		for _, density := range []float32{1, 2} {
			name := map[theme.Scheme]string{theme.Light: "light", theme.Dark: "dark"}[scheme] + "-" + map[float32]string{1: "1x", 2: "2x"}[density]
			t.Run(name, func(t *testing.T) {
				got := drawFooter(t, scheme, density)
				path := filepath.Join("testdata", "native-line-"+name+".png")
				if *update {
					if err := os.MkdirAll("testdata", 0o755); err != nil {
						t.Fatal(err)
					}
					file, err := os.Create(path)
					if err != nil {
						t.Fatal(err)
					}
					defer file.Close()
					if err := png.Encode(file, got); err != nil {
						t.Fatal(err)
					}
					return
				}
				file, err := os.Open(path)
				if err != nil {
					t.Fatalf("%s has no approved picture: draw it with -update and look at it before committing", name)
				}
				defer file.Close()
				want, err := png.Decode(file)
				if err != nil {
					t.Fatal(err)
				}
				if got.Bounds() != want.Bounds() {
					t.Fatalf("%s is %v, the approved picture is %v", name, got.Bounds(), want.Bounds())
				}
				changed := 0
				for y := got.Bounds().Min.Y; y < got.Bounds().Max.Y; y++ {
					for x := got.Bounds().Min.X; x < got.Bounds().Max.X; x++ {
						if far(got.At(x, y), want.At(x, y)) {
							changed++
						}
					}
				}
				if changed > 0 {
					failed := strings.TrimSuffix(path, ".png") + ".failed.png"
					if file, err := os.Create(failed); err == nil {
						_ = png.Encode(file, got)
						file.Close()
					}
					t.Errorf("%s changed in %d pixels; the drawn picture is %s", name, changed, failed)
				}
			})
		}
	}
}

// far reports whether two pixels differ by more than 16 of 255 in a channel:
// the software renderer a build server uses and a machine's GPU disagree by
// less than that on an edge, and by whole colours on a real change.
func far(a, b color.Color) bool {
	ar, ag, ab, aa := a.RGBA()
	br, bg, bb, ba := b.RGBA()
	d := func(x, y uint32) bool {
		if x > y {
			x, y = y, x
		}
		return (y-x)>>8 > 16
	}
	return d(ar, br) || d(ag, bg) || d(ab, bb) || d(aa, ba)
}

// The marks and the words are drawn in an ink that reads against the ground in
// both schemes, and the strokes come out solid and not smeared: at least a
// share of the ink pixels reach the full contrast.
func TestTheNativeLineIsLegibleInBothSchemes(t *testing.T) {
	for _, scheme := range []theme.Scheme{theme.Light, theme.Dark} {
		for _, density := range []float32{1, 2, 3} {
			got := drawFooter(t, scheme, density)
			ground := theme.New(scheme).Colours.Background
			var best float64
			var ink, solid int
			ratios := []float64{}
			for y := 0; y < got.Bounds().Dy(); y++ {
				for x := 0; x < got.Bounds().Dx(); x++ {
					r := contrast(got.RGBAAt(x, y), ground)
					ratios = append(ratios, r)
					best = max(best, r)
				}
			}
			for _, r := range ratios {
				if r > 1.5 {
					ink++
				}
				if r >= 4.5 {
					solid++
				}
			}
			if best < 4.5 {
				t.Errorf("%v at %vx: the best contrast is %.1f, want 4.5 at least", scheme, density, best)
			}
			if ink == 0 || float64(solid)/float64(ink) < 0.15 {
				t.Errorf("%v at %vx: %d of %d ink pixels read at 4.5, the strokes are smeared", scheme, density, solid, ink)
			}
		}
	}
}

func contrast(a color.RGBA, b color.NRGBA) float64 {
	la := luminance(float64(a.R), float64(a.G), float64(a.B))
	lb := luminance(float64(b.R), float64(b.G), float64(b.B))
	if la < lb {
		la, lb = lb, la
	}
	return (la + 0.05) / (lb + 0.05)
}

func luminance(r, g, b float64) float64 {
	f := func(v float64) float64 {
		v /= 255
		if v <= 0.03928 {
			return v / 12.92
		}
		return math.Pow((v+0.055)/1.055, 2.4)
	}
	return 0.2126*f(r) + 0.7152*f(g) + 0.0722*f(b)
}
