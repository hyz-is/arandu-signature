package native

import (
	"crypto/ed25519"
	"encoding/base64"
	"net/url"
	"time"

	"github.com/hyz-is/arandu-signature/internal/licence"
)

// HyzisPublicKey verifies every partnership licence, as the web half does with
// the same key: this package cannot import that one without linking the web
// framework into the app, so the key is written twice and a test of the module
// keeps the two equal.
const HyzisPublicKey = "FRKr0UBvijZ5g2XtQbsZOGGZtbkrvngjQc4+j1/qr+Q="

var publicKey = func() ed25519.PublicKey {
	key, err := base64.StdEncoding.DecodeString(HyzisPublicKey)
	if err != nil || len(key) != ed25519.PublicKeySize {
		panic("signature/native: the embedded licence key is not an Ed25519 public key")
	}
	return ed25519.PublicKey(key)
}()

// Licensed reports whether token is a valid partnership licence for host
// today, and what it says.
func Licensed(token, host string) (licence.Claims, bool) {
	claims, err := licence.Verify(token, publicKey, hostOf(host), time.Now())
	return claims, err == nil
}

// hostOf is the host of an address, or the address itself when it has no
// scheme.
func hostOf(address string) string {
	if u, err := url.Parse(address); err == nil && u.Host != "" {
		return u.Host
	}
	return address
}
