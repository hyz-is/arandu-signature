package signature

import (
	"fmt"
	"net/url"
)

// The two addresses the marks link to.
const (
	DefaultHyzisURL  = "https://hyz.is"
	DefaultAranduURL = "https://arandu.io"
)

// The languages the words of the line come in.
const (
	// PortugueseBR reads "Desenvolvido por … em …", and is the default.
	PortugueseBR = "pt-BR"
	// English reads "Built by … with …".
	English = "en"
)

// Config is how a project signs.
//
// Enabled is the contract: the signature appears only on a product whose client
// agreed to the brand partnership, and a project that has not says nothing,
// so the line renders as nothing at all rather than as a hidden element.
type Config struct {
	// Enabled turns the signature on. Zero is off.
	Enabled bool
	// Lang is PortugueseBR or English. Empty is PortugueseBR.
	Lang string
	// HyzisURL and AranduURL are where the marks link to. Empty is the
	// default address of each; a value is refused unless it is an absolute
	// https address.
	HyzisURL  string
	AranduURL string

	// License is the partnership licence HYZIS issued for this project. With
	// a valid one for AppURL's host, the signature is always on: Enabled is
	// ignored, and Middleware puts the line back on any page that lost it.
	//
	// Keep it a Go constant in bootstrap/app.go rather than an environment
	// variable: removing it is then a change to the source, reviewed and
	// recorded in the repository like any other, not a setting flipped on a
	// server. An invalid or expired licence is logged and ignored; it never
	// stops the application.
	License string
	// AppURL is the address the application is served at (APP_URL), whose
	// host the licence has to name.
	AppURL string
}

// Validate refuses a configuration that would render a broken or unsafe link.
func (c Config) Validate() error {
	if c.Lang != "" && c.Lang != PortugueseBR && c.Lang != English {
		return fmt.Errorf("signature: Config.Lang is %q, and has to be %q or %q", c.Lang, PortugueseBR, English)
	}
	for name, address := range map[string]string{"HyzisURL": c.HyzisURL, "AranduURL": c.AranduURL} {
		if address == "" {
			continue
		}
		u, err := url.Parse(address)
		if err != nil || u.Scheme != "https" || u.Host == "" {
			return fmt.Errorf("signature: Config.%s is %q, and has to be an absolute https address", name, address)
		}
	}
	return nil
}

func (c Config) withDefaults() Config {
	if c.Lang == "" {
		c.Lang = PortugueseBR
	}
	if c.HyzisURL == "" {
		c.HyzisURL = DefaultHyzisURL
	}
	if c.AranduURL == "" {
		c.AranduURL = DefaultAranduURL
	}
	return c
}

// Words are the two pieces of text around the marks, in the language asked.
func Words(lang string) (by, with string) {
	if lang == English {
		return "Built by", "with"
	}
	return "Desenvolvido por", "em"
}
