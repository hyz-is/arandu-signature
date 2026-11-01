# Arandu Signature

The line HYZIS signs the products it develops with, under the brand
partnership: **"Desenvolvido por [HYZIS] em [Arandu]"**, each mark a link to its
site (https://hyz.is, https://arandu.io), drawn in the ink of the screen around
it. One module for both halves of an [Arandu](https://arandu.io) product:

- `github.com/hyz-is/arandu-signature` — the web half, a line for a
  [Kyse](https://github.com/arandu-io/kyse) layout;
- `github.com/hyz-is/arandu-signature/native` — the native half, for an
  [Ayra](https://github.com/arandu-io/ayra) app: the line and the About screen
  it belongs on, with the marks opening the phone's browser.

The signature appears only when the client's contract allows it: `Enabled` is
zero by default, and a project that does not sign renders nothing at all.

```sh
go get github.com/hyz-is/arandu-signature@latest
```

## Where it goes

- **Web:** one line in the footer of the public layout. Not in the panel's
  chrome, the sign-in form or a transactional screen.
- **Native app:** on an About screen ("Sobre", reached from the account screen),
  never next to signing out.

## Web: a Kyse layout

Build the module in `bootstrap/app.go`, with the contract in the project's
configuration, and register it with the others:

```go
sig, err := signature.New(signature.Config{Enabled: cfg.App.SignatureEnabled})
if err != nil {
	return err
}
k.Register(sig)
```

Then draw the line where the footer ends, in the layout:

```go
{!! signature.Kyse(signature.Props{Enabled: true, Class: "footer-signature"}) !!}
```

or hand the module's `sig.HTML()` to the layout's data. The line brings its own
stylesheet and the two marks as view assets served from the origin, so the
layout adds nothing else and a strict Content-Security-Policy holds: no inline
style, no script. `Lang: signature.English` reads "Built by … with …";
`HyzisURL` and `AranduURL` take another https address.

## Native: an Ayra app

```go
import sig "github.com/hyz-is/arandu-signature/native"

var about sig.State

func layoutAbout(c ayra.Context) ayra.Dimensions {
	return sig.About(c, &about, sig.AboutProps{
		Mark: brandmark, Name: "Nivela", Version: "0.1.3",
		Description: "Acompanhe orçamentos, visitas e o histórico de cada serviço.",
		Signature:   sig.Options{Enabled: true},
	})
}
```

`sig.Line(c, &state, sig.Options{Enabled: true})` draws the line alone. Each
mark is a target of at least 40 dp; a tap calls `sig.OpenURL`, which opens the
system browser: an `ACTION_VIEW` intent on Android, a new tab on the browser
build (js/wasm), the desktop's opener on macOS, Windows and Linux. On iOS it
does nothing yet (Ayra exposes no handle to `openURL`), and the marks stay a
visible credit there.

The native half targets Ayra v0.2.0 and imports nothing from the web half.

## Gates

```sh
export GOWORK=off
gofmt -l .
go build ./...
go vet ./...
GOOS=js GOARCH=wasm go vet ./native/
go test -race ./...
```

## License

MIT. See [LICENSE.md](LICENSE.md). The HYZIS and Arandu marks are the brands of
their owners and are shipped for the signature only.
