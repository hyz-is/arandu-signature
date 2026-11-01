package signature

import (
	"crypto/ed25519"
	"encoding/base64"
	"time"

	"github.com/hyz-is/arandu-signature/internal/licence"
)

// hyzisPublicKey verifies every partnership licence. Only HYZIS holds the
// private half, so only HYZIS issues a licence; the key is public on purpose.
// HyzisPublicKey is the same key, exported so a test can hold the native copy to it.
const HyzisPublicKey = hyzisPublicKey

const hyzisPublicKey = "FRKr0UBvijZ5g2XtQbsZOGGZtbkrvngjQc4+j1/qr+Q="

// PublicKey is HYZIS's licence key, decoded once.
var PublicKey = func() ed25519.PublicKey {
	key, err := base64.StdEncoding.DecodeString(hyzisPublicKey)
	if err != nil || len(key) != ed25519.PublicKeySize {
		panic("signature: the embedded licence key is not an Ed25519 public key")
	}
	return ed25519.PublicKey(key)
}()

// Claims is what a partnership licence says: the project, the hosts it is
// valid on, and until when.
type Claims = licence.Claims

// Verify reads a partnership licence for host. An empty host checks only the
// signature and the expiry; localhost and loopback addresses match any
// licence.
func Verify(token, host string, now time.Time) (Claims, error) {
	return licence.Verify(token, licenceKey, host, now)
}

// licenceKey is the key Verify checks with: PublicKey, and in this package's
// own tests a key they generate, since HYZIS's private key is in no test.
var licenceKey = PublicKey
