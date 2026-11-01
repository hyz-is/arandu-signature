// Package signature is the line HYZIS signs the products it develops with,
// under the brand partnership: "Desenvolvido por [HYZIS] em [Arandu]", each
// mark a link, drawn in the ink of the page around it.
//
// This package is the web half, for a Kyse layout. The native half, for an
// Ayra app, is the package github.com/hyz-is/arandu-signature/native, so a
// project with no native target links no native code.
package signature

import (
	"html/template"
	"log/slog"
	"net/url"
	"sync/atomic"
	"time"

	"github.com/arandu-io/framework/foundation"
	fhttp "github.com/arandu-io/framework/http"
)

// Module is the signature as an application registers it.
type Module struct {
	cfg      Config
	licensed bool
	claims   Claims
}

var _ foundation.Module = (*Module)(nil)

// forced is the configuration of a licensed project, read by Kyse: a layout
// that asks for the line with Enabled false still gets it.
var forced atomic.Pointer[Config]

// New builds the module from the project's configuration. It refuses a
// configuration that would render a broken or unsafe link, and never a
// licence: an invalid or expired one is logged and the project signs as its
// Enabled says.
func New(cfg Config) (*Module, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	m := &Module{cfg: cfg.withDefaults()}
	forced.Store(nil)
	if cfg.License != "" {
		claims, err := Verify(cfg.License, hostOf(cfg.AppURL), time.Now())
		if err != nil {
			slog.Warn("signature: the partnership licence is not valid here and is ignored", "error", err)
		} else {
			m.licensed, m.claims = true, claims
			m.cfg.Enabled = true
			signed := m.cfg
			forced.Store(&signed)
			slog.Info("signature: partnership licence for " + claims.Project + " until " + claims.ExpiresAt().Format(time.DateOnly))
		}
	}
	return m, nil
}

// hostOf is the host of an address, or the address itself when it has no
// scheme.
func hostOf(address string) string {
	if u, err := url.Parse(address); err == nil && u.Host != "" {
		return u.Host
	}
	return address
}

// Name is the module identifier.
func (m *Module) Name() string { return "signature" }

// Routes registers nothing: the marks and the stylesheet are view assets the
// framework already serves.
func (m *Module) Routes(*fhttp.Router) {}

// Enabled reports whether this project signs.
func (m *Module) Enabled() bool { return m.cfg.Enabled }

// Licensed reports whether a valid partnership licence forces the signature
// on, and what it says.
func (m *Module) Licensed() (Claims, bool) { return m.claims, m.licensed }

// HTML is the line for this project, or nothing when it does not sign.
func (m *Module) HTML() template.HTML {
	return Kyse(m.props())
}

func (m *Module) props() Props {
	return Props{
		Enabled:   m.cfg.Enabled,
		Lang:      m.cfg.Lang,
		HyzisURL:  m.cfg.HyzisURL,
		AranduURL: m.cfg.AranduURL,
	}
}
