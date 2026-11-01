package native

import (
	"testing"

	"github.com/arandu-io/ayra/engine/unit"
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
