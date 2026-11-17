---
name: signature-package
description: Use when installing or changing hyz-is/arandu-signature, the HYZIS partnership signature for Kyse layouts and Ayra apps.
---

# Working on the signature package

The package draws "Desenvolvido por [HYZIS] em [Arandu]" with each mark a link.
It owns no table, no route and no state: the web half registers three view
assets (two marks and a stylesheet) and renders one line; the native half
(`native/`) draws the line and an About screen for an Ayra app and opens the
system browser.

## Installing it in an application

Composition is explicit, in `App` in `bootstrap/app.go`, once
`NewAppServiceProvider` has returned. `Configure` has no way to return the error
of `New`, so the module is built where the error can leave, and listed in the
`WithProviders(...)` chain the builder is made of:

```go
sig, err := signature.New(signature.Config{Enabled: contractAllowsIt})
if err != nil {
	_ = drivers.Close()
	return nil, err
}

return Configure(cfg, drivers, app).
	WithProviders(sig).
	WithVersion(version, commit).
	Create()
```

The layout draws `{!! signature.Kyse(signature.Props{Enabled: true}) !!}` in
the public footer. A letter draws `signature.Email(signature.EmailProps{Enabled: true, Origin: ...})`
in its foot, with `signature.EmailStyle()` in its `<style>` and
`signature.EmailText(...)` in the text part: PNG marks the site serves, in
inline styles, because a mail client draws no SVG and no mask. An Ayra app calls `native.About` on its About screen, reached
from the account screen, never beside signing out.

## Rules

- `Enabled` is the client's contract and is false by default; a disabled line
  renders nothing.
- The marks are masks in the surrounding ink; no inline style or script, so a
  strict CSP holds.
- `native` never imports the root package: an app links no web framework.
- A link is an absolute https address or it is refused.
