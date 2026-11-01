package signature

import (
	"bytes"
	"net/http"
	"strconv"
	"strings"

	"github.com/arandu-io/hesape/view"
)

// Middleware puts the line back on a page that lost it, when the project
// holds a partnership licence. Install it once, with the application's
// middleware:
//
//	m.Append(sig.Middleware())
//
// It reads only what it may change: a whole HTML page (text/html, status 200,
// not an htmx fragment, not already compressed). A page whose body carries the
// line's Marker is left exactly as it was; one without it gets the line before
// its closing </body>. A page that signs but no longer links the line's
// stylesheet -- its layout drew the line with HeadStylesheet and the <head>
// lost the link -- gets the link back before </head>. Without a licence it is
// a pass-through.
func (m *Module) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if !m.licensed {
			return next
		}
		block := []byte(`<div class="arandu-signature-restored" data-arandu-signature-restored>` + string(Kyse(m.props())) + `</div>`)
		sheet := []byte(Stylesheet())
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet || r.Header.Get("HX-Request") == "true" {
				next.ServeHTTP(w, r)
				return
			}
			rw := &restoringWriter{ResponseWriter: w}
			next.ServeHTTP(rw, r)
			rw.finish(block, sheet)
		})
	}
}

// restoringWriter holds back a whole HTML page and passes everything else
// through untouched.
type restoringWriter struct {
	http.ResponseWriter
	decided   bool
	buffering bool
	status    int
	body      bytes.Buffer
}

func (rw *restoringWriter) decide(status int) {
	if rw.decided {
		return
	}
	rw.decided = true
	rw.status = status
	h := rw.Header()
	rw.buffering = status == http.StatusOK &&
		strings.HasPrefix(strings.ToLower(h.Get("Content-Type")), "text/html") &&
		h.Get("Content-Encoding") == ""
	if !rw.buffering {
		rw.ResponseWriter.WriteHeader(status)
	}
}

func (rw *restoringWriter) WriteHeader(status int) { rw.decide(status) }

func (rw *restoringWriter) Write(b []byte) (int, error) {
	rw.decide(http.StatusOK)
	if !rw.buffering {
		return rw.ResponseWriter.Write(b)
	}
	return rw.body.Write(b)
}

// Flush gives up on a page that streams: it is sent as it is from here on.
func (rw *restoringWriter) Flush() {
	if rw.buffering {
		rw.buffering = false
		rw.ResponseWriter.WriteHeader(rw.status)
		_, _ = rw.ResponseWriter.Write(rw.body.Bytes())
		rw.body.Reset()
	}
	if f, ok := rw.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// Unwrap lets http.ResponseController reach the writer underneath.
func (rw *restoringWriter) Unwrap() http.ResponseWriter { return rw.ResponseWriter }

// finish sends the held page, with the line put back when it lacks it.
func (rw *restoringWriter) finish(block, sheet []byte) {
	if !rw.decided || !rw.buffering {
		return
	}
	page := rw.body.Bytes()
	if !bytes.Contains(page, []byte(Marker)) {
		page = insertBefore(page, "</body>", block)
	} else if !bytes.Contains(page, []byte(view.Asset(StylesheetAsset))) {
		page = insertBefore(page, "</head>", sheet)
	}
	rw.Header().Set("Content-Length", strconv.Itoa(len(page)))
	rw.ResponseWriter.WriteHeader(rw.status)
	_, _ = rw.ResponseWriter.Write(page)
}

// insertBefore puts add right before the last occurrence of the closing tag,
// matched without regard to case, and leaves a page without the tag as it was.
func insertBefore(page []byte, tag string, add []byte) []byte {
	at := bytes.LastIndex(bytes.ToLower(page), []byte(tag))
	if at < 0 {
		return page
	}
	out := make([]byte, 0, len(page)+len(add))
	out = append(out, page[:at]...)
	out = append(out, add...)
	return append(out, page[at:]...)
}
