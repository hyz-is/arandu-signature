// Package licence is the partnership licence HYZIS issues to a project whose
// client agreed to carry the signature: who the project is, the hosts it is
// served from, and until when, signed with HYZIS's Ed25519 key.
//
// It holds no state and reads no file; the web half and the native half of the
// module both verify with it, so it imports the standard library only.
package licence

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"
)

// Claims is what a licence says.
type Claims struct {
	// Project names the product, for the boot line and for whoever reads it.
	Project string `json:"p"`
	// Domains are the hosts the licence is valid on, without port or scheme.
	Domains []string `json:"d"`
	// Issued and Expires are Unix seconds.
	Issued  int64 `json:"i"`
	Expires int64 `json:"e"`
}

// ExpiresAt is when the licence stops being valid.
func (c Claims) ExpiresAt() time.Time { return time.Unix(c.Expires, 0).UTC() }

// The refusals of Verify. A caller tells them apart to log why; none of them
// is ever a reason to stop the application.
var (
	ErrMalformed = errors.New("licence: not a licence")
	ErrSignature = errors.New("licence: the signature does not match HYZIS's key")
	ErrExpired   = errors.New("licence: expired")
	ErrHost      = errors.New("licence: not issued for this host")
)

var encoding = base64.RawURLEncoding

// Sign issues a licence: base64url(claims) "." base64url(signature).
func Sign(c Claims, key ed25519.PrivateKey) (string, error) {
	payload, err := json.Marshal(c)
	if err != nil {
		return "", err
	}
	body := encoding.EncodeToString(payload)
	return body + "." + encoding.EncodeToString(ed25519.Sign(key, []byte(body))), nil
}

// Verify reads a licence, checks its signature with key, its expiry against
// now and, when host is not empty, that it was issued for host. A licence
// matches any host on a developer's machine (localhost and loopback
// addresses), so a project runs the same way locally as in production.
func Verify(token string, key ed25519.PublicKey, host string, now time.Time) (Claims, error) {
	body, sig, ok := strings.Cut(strings.TrimSpace(token), ".")
	if !ok || body == "" || sig == "" {
		return Claims{}, ErrMalformed
	}
	signature, err := encoding.DecodeString(sig)
	if err != nil || len(signature) != ed25519.SignatureSize {
		return Claims{}, ErrMalformed
	}
	if len(key) != ed25519.PublicKeySize || !ed25519.Verify(key, []byte(body), signature) {
		return Claims{}, ErrSignature
	}
	payload, err := encoding.DecodeString(body)
	if err != nil {
		return Claims{}, ErrMalformed
	}
	var c Claims
	if err := json.Unmarshal(payload, &c); err != nil || c.Project == "" || c.Expires == 0 {
		return Claims{}, ErrMalformed
	}
	if !now.Before(c.ExpiresAt()) {
		return c, fmt.Errorf("%w on %s", ErrExpired, c.ExpiresAt().Format(time.DateOnly))
	}
	if host != "" && !c.Covers(host) {
		return c, fmt.Errorf("%w %q", ErrHost, host)
	}
	return c, nil
}

// Covers reports whether the licence is valid on host: one of its domains, or
// a developer's own machine. A port is ignored.
func (c Claims) Covers(host string) bool {
	host = strings.ToLower(strings.TrimSpace(host))
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}
	host = strings.Trim(host, "[]")
	if host == "localhost" || strings.HasSuffix(host, ".localhost") {
		return true
	}
	if ip := net.ParseIP(host); ip != nil && ip.IsLoopback() {
		return true
	}
	for _, d := range c.Domains {
		if strings.EqualFold(strings.TrimSpace(d), host) {
			return true
		}
	}
	return false
}
