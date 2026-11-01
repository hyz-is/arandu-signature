//go:build ios

package native

// OpenURL does nothing on iOS yet: Ayra exposes no handle to UIApplication's
// openURL, and the mark stays a visible credit without a link there.
func OpenURL(url string) error { return nil }
