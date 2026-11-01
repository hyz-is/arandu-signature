package unit_test

import (
	"testing"

	signature "github.com/hyz-is/arandu-signature"
	"github.com/hyz-is/arandu-signature/native"
)

// TestBothHalvesVerifyWithTheSameKey keeps the native copy of HYZIS's licence
// key equal to the web one: a licence valid on the site has to be valid in the
// app.
func TestBothHalvesVerifyWithTheSameKey(t *testing.T) {
	if native.HyzisPublicKey != signature.HyzisPublicKey {
		t.Fatal("the native package verifies licences with a different key than the web package")
	}
}
