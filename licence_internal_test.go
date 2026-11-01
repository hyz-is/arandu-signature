package signature

import (
	"crypto/ed25519"
	"crypto/rand"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/hyz-is/arandu-signature/internal/licence"
)

// issue signs a licence with a key of the test's own, and makes the package
// verify with it for the test's duration.
func issue(t *testing.T, claims licence.Claims) string {
	t.Helper()
	public, private, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	previous := licenceKey
	licenceKey = public
	t.Cleanup(func() { licenceKey = previous; forced.Store(nil) })
	token, err := licence.Sign(claims, private)
	if err != nil {
		t.Fatal(err)
	}
	return token
}

func valid(domains ...string) licence.Claims {
	now := time.Now()
	return licence.Claims{Project: "demo", Domains: domains, Issued: now.Unix(), Expires: now.AddDate(10, 0, 0).Unix()}
}

func TestALicensedProjectSignsEvenWhenItAsksNotTo(t *testing.T) {
	token := issue(t, valid("demo.example"))
	m, err := New(Config{Enabled: false, License: token, AppURL: "https://demo.example"})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := m.Licensed(); !ok || !m.Enabled() {
		t.Fatalf("a valid licence did not force the signature on")
	}
	if got := string(Kyse(Props{Enabled: false})); !strings.Contains(got, Marker) {
		t.Fatalf("a layout asking for no line drew nothing under a licence: %q", got)
	}
}

func TestALicenceForAnotherHostIsIgnored(t *testing.T) {
	token := issue(t, valid("demo.example"))
	m, err := New(Config{Enabled: false, License: token, AppURL: "https://other.example"})
	if err != nil {
		t.Fatalf("a licence for another host stopped the module: %v", err)
	}
	if _, ok := m.Licensed(); ok {
		t.Fatal("a licence for another host was accepted")
	}
	if got := Kyse(Props{Enabled: false}); got != "" {
		t.Fatalf("an unlicensed, disabled project drew %q", got)
	}
}

func TestAnExpiredOrTamperedLicenceIsIgnored(t *testing.T) {
	expired := valid("demo.example")
	expired.Expires = time.Now().Add(-time.Hour).Unix()
	for name, token := range map[string]string{
		"expired":  issue(t, expired),
		"tampered": issue(t, valid("demo.example")) + "x",
		"garbage":  "not-a-licence",
	} {
		m, err := New(Config{License: token, AppURL: "https://demo.example"})
		if err != nil {
			t.Fatalf("%s: the licence stopped the module: %v", name, err)
		}
		if _, ok := m.Licensed(); ok {
			t.Errorf("%s licence was accepted", name)
		}
	}
}

func TestLocalhostMatchesAnyLicence(t *testing.T) {
	token := issue(t, valid("demo.example"))
	m, _ := New(Config{License: token, AppURL: "http://localhost:8080"})
	if _, ok := m.Licensed(); !ok {
		t.Fatal("a licence did not cover the developer's machine")
	}
}

func serve(t *testing.T, m *Module, contentType, body string, header map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	handler := m.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", contentType)
		w.Header().Set("Content-Length", strconv.Itoa(len(body)))
		for k, v := range header {
			w.Header().Set(k, v)
		}
		_, _ = w.Write([]byte(body))
	}))
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	if header["HX-Request"] != "" {
		r.Header.Set("HX-Request", header["HX-Request"])
	}
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)
	return w
}

func licensed(t *testing.T) *Module {
	t.Helper()
	m, err := New(Config{License: issue(t, valid("demo.example")), AppURL: "https://demo.example"})
	if err != nil {
		t.Fatal(err)
	}
	return m
}

func TestThePageThatLostTheLineGetsItBack(t *testing.T) {
	w := serve(t, licensed(t), "text/html; charset=utf-8", "<html><body><main>x</main></body></html>", nil)
	body := w.Body.String()
	if strings.Count(body, `data-arandu-signature `) != 1 || !strings.Contains(body, "data-arandu-signature-restored") {
		t.Fatalf("the line was not put back: %s", body)
	}
	if i, j := strings.Index(body, "data-arandu-signature-restored"), strings.Index(body, "</body>"); i < 0 || i > j {
		t.Fatal("the line was not put before </body>")
	}
	if got := w.Header().Get("Content-Length"); got != strconv.Itoa(len(body)) {
		t.Fatalf("Content-Length is %s for a body of %d", got, len(body))
	}
}

func TestThePageThatSignsIsLeftAsItIs(t *testing.T) {
	m := licensed(t)
	page := "<html><body><footer>" + string(Kyse(Props{})) + "</footer></body></html>"
	if got := serve(t, m, "text/html", page, nil).Body.String(); got != page {
		t.Fatalf("a page that signs was changed:\n%s", got)
	}
}

func TestWhatIsNotAWholePageIsLeftAlone(t *testing.T) {
	m := licensed(t)
	page := "<html><body>x</body></html>"
	for name, c := range map[string]struct {
		ct     string
		header map[string]string
	}{
		"json":       {"application/json", nil},
		"htmx":       {"text/html", map[string]string{"HX-Request": "true"}},
		"compressed": {"text/html", map[string]string{"Content-Encoding": "gzip"}},
	} {
		if got := serve(t, m, c.ct, page, c.header).Body.String(); got != page {
			t.Errorf("%s: the body was changed: %s", name, got)
		}
	}
}

func TestWithoutALicenceTheMiddlewarePassesThrough(t *testing.T) {
	m, _ := New(Config{Enabled: true})
	page := "<html><body>x</body></html>"
	if got := serve(t, m, "text/html", page, nil).Body.String(); got != page {
		t.Fatalf("an unlicensed project's page was changed: %s", got)
	}
}

func TestTheStylesheetResistsHidingRules(t *testing.T) {
	css := string(stylesheet("/h.svg", "/a.svg"))
	for _, rule := range []string{
		"display: inline-flex !important", "visibility: visible !important", "opacity: 1 !important",
		"position: static !important", "clip-path: none !important", "transform: none !important",
		"filter: none !important", "font-size: max(12px, 1em) !important", "p.arandu-signature[data-arandu-signature]",
	} {
		if !strings.Contains(css, rule) {
			t.Errorf("the stylesheet lacks %q", rule)
		}
	}
}

// A page that still signs but lost the stylesheet link its layout put in the
// <head> gets the link back there, once.
func TestThePageThatLostItsStylesheetGetsItBack(t *testing.T) {
	page := `<html><head><title>x</title></head><body><p ` + Marker + `>line</p></body></html>`
	out := string(insertBefore([]byte(page), "</head>", []byte(Stylesheet())))
	if !strings.Contains(out, string(Stylesheet())+"</head>") {
		t.Fatalf("the stylesheet did not go back into the head:\n%s", out)
	}
	if got := insertBefore([]byte("<p>no head</p>"), "</head>", []byte("x")); string(got) != "<p>no head</p>" {
		t.Fatalf("a page without the tag changed: %s", got)
	}
}
