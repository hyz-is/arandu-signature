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

	"github.com/arandu-io/framework/foundation"
	fhttp "github.com/arandu-io/framework/http"
)

// Module is the signature as an application registers it.
type Module struct {
	cfg Config
}

var _ foundation.Module = (*Module)(nil)

// New builds the module from the project's configuration. It refuses a
// configuration that would render a broken or unsafe link.
func New(cfg Config) (*Module, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &Module{cfg: cfg.withDefaults()}, nil
}

// Name is the module identifier.
func (m *Module) Name() string { return "signature" }

// Routes registers nothing: the marks and the stylesheet are view assets the
// framework already serves.
func (m *Module) Routes(*fhttp.Router) {}

// Enabled reports whether this project signs.
func (m *Module) Enabled() bool { return m.cfg.Enabled }

// HTML is the line for this project, or nothing when it does not sign.
func (m *Module) HTML() template.HTML {
	return Kyse(Props{
		Enabled:   m.cfg.Enabled,
		Lang:      m.cfg.Lang,
		HyzisURL:  m.cfg.HyzisURL,
		AranduURL: m.cfg.AranduURL,
	})
}
