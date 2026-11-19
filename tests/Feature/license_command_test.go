package feature_test

import (
	"encoding/base64"
	"encoding/json"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestTheLicenceCommandKeepsEveryDomain signs with -domain given twice, one of
// them listing two hosts: every host is in the licence. A plain string flag
// kept only the last -domain and signed a licence the site's main host failed.
func TestTheLicenceCommandKeepsEveryDomain(t *testing.T) {
	key := filepath.Join(t.TempDir(), "signature-license.key")
	run := func(args ...string) string {
		t.Helper()
		cmd := exec.Command("go", append([]string{"run", "../../cmd/signature-license"}, args...)...)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("signature-license %v: %v\n%s", args, err, out)
		}
		return strings.TrimSpace(string(out))
	}
	run("-genkey", "-key", key)
	token := run("-key", key, "-project", "example", "-domain", "example.test", "-domain", "www.example.test, App.Example.test", "-years", "1")

	payload, err := base64.RawURLEncoding.DecodeString(strings.SplitN(token, ".", 2)[0])
	if err != nil {
		t.Fatalf("the licence payload does not decode: %v", err)
	}
	var claims struct {
		Domains []string `json:"d"`
	}
	if err := json.Unmarshal(payload, &claims); err != nil {
		t.Fatal(err)
	}
	want := []string{"example.test", "www.example.test", "app.example.test"}
	if strings.Join(claims.Domains, ",") != strings.Join(want, ",") {
		t.Fatalf("the licence holds %v, want %v", claims.Domains, want)
	}
}
