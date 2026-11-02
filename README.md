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
  never next to signing out. Where the product is a showcase of HYZIS, the app
  may also put the line at the foot of the sign-in screen: it is one widget,
  `sig.LineProps{Enabled: true}.Layout`, that goes on any screen.

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

The line alone, on any screen, is `sig.LineProps` (the same fields as
`sig.Options`) with the `Layout` every Ayra widget has. It is centred across the
room it is given, in the ink of the screen around it, and takes no room when
the project does not sign. At the foot of a sign-in screen:

```go
var signed sig.State

func layoutLogin(c ayra.Context) ayra.Dimensions {
	return layout.Flex{Axis: layout.Vertical}.Layout(c.Context,
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions { return form(c.With(gtx)) }),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return sig.LineProps{Enabled: true}.Layout(c.With(gtx), &signed)
		}),
	)
}
```

One `sig.State` for each place the line is drawn. `sig.Line(c, &state,
sig.Options{Enabled: true})` is the same line at its own width, not centred.
Each mark answers a press within 44 dp (Ayra's `TouchArea`) although the figure
is smaller, and is drawn on the pixels it covers, so its edges are sharp at any
density, in both schemes. A tap calls `sig.OpenURL`, which opens the
system browser: an `ACTION_VIEW` intent on Android, a new tab on the browser
build (js/wasm), the desktop's opener on macOS, Windows and Linux. On iOS it
does nothing yet (Ayra exposes no handle to `openURL`), and the marks stay a
visible credit there.

The native half targets Ayra v0.9.1 and imports nothing from the web half.

## Partnership licence

A project whose contract keeps the brand for a period gets a licence HYZIS
signs (Ed25519) for its domain and its term. With a valid one, the signature
is always on: `Enabled` is ignored, there is no option that hides it, and the
boot log says `signature: partnership licence for <project> until <date>`.

```go
// bootstrap/app.go -- a constant in the source, not an environment variable:
// removing it is a reviewed change to the repository, not a server setting.
const signatureLicense = "eyJw…"

sig, err := signature.New(signature.Config{License: signatureLicense, AppURL: cfg.App.URL})
k.Register(sig)
// with the application's middleware: puts the line back on a page that lost it
m.Append(sig.Middleware())
```

`Middleware` reads only whole HTML pages (status 200, not htmx fragments, not
compressed) and adds the line before `</body>` when the page lacks its marker.
The stylesheet keeps the line visible against common hiding rules
(`!important` on display, visibility, opacity, position, size, clip and
transform). In an app, `sig.Options{License: …, Host: serverURL}` draws it
always on the About screen. An invalid or expired licence is logged and
ignored; it never stops the application. `localhost` and loopback addresses
match any licence.

This is friction, not a lock: the client owns the source, and a developer who
deletes the constant, the middleware or the component removes the signature.
What the licence changes is that removing it takes a deliberate edit to the
code, recorded in the repository.

HYZIS issues licences with the command in this module, using its private key
(kept outside every repository):

```sh
go run github.com/hyz-is/arandu-signature/cmd/signature-license \
  -key ~/.config/hyzis/signature-license.key -project <name> -domain <host> -years 10
```

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

## Since v0.3.0

- **The line on any screen.** `sig.LineProps{...}.Layout(c, &state)` draws the
  signature centred across the room it is given; the sign-in screen's foot is
  one place, About is another.
- **Ayra v0.9.1.** Marks answer within 44 dp and are resampled to the pixels
  they cover.

## Since v0.2.1

- **Stylesheet in the head.** Add `{!! signature.Stylesheet() !!}` to the
  layout's `<head>` and draw the line with
  `signature.Kyse(signature.Props{Enabled: true, HeadStylesheet: true})`.
- **Mark size.** `Props.Size: signature.SizeSmall` or `SizeLarge` on the web,
  or `--arandu-signature-scale` from the layout's own stylesheet;
  `Options.Height` (dp) in an app.
- **About screen.** `AboutProps.Back` and `AboutProps.Eyebrow` draw the app's
  own top bar; `Placement: sig.AfterContent` centres the signature right after
  the description instead of the foot; leave `Name` empty when `Mark` already
  carries the wordmark.
- **iOS.** `OpenURL` opens the address through UIKit.
