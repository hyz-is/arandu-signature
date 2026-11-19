# Upgrade Guide

## Unreleased

## v0.5.2

Nothing to change in code. Name the file that holds the licence
`config/signature.go`, after the module, as the README now shows.

## v0.5.1

Nothing to change in code. A project that keeps its licence elsewhere (in
`bootstrap/app.go` or a package of its own) may move the constant to
`config/partnership.go`, as the README now shows. A licence signed with two
`-domain` flags by v0.5.0 holds only the last host: sign it again.

## v0.5.0

Nothing to change; nothing that compiled stops compiling. To sign a letter's foot,
draw `signature.Email(signature.EmailProps{Enabled: true, Origin: "https://your.site"})`
in the layout and add `signature.EmailStyle()` to its `<style>`.

## v0.4.1

Nothing to change. The module requires hesape v0.109.0 and framework v0.71.0.

## v0.4.0

Nothing to change. `Middleware` leaves a request alone when it carries
`YPY-Request: true`, where it used to read the header of the old library, so a
page fetched by the script is still not rewritten and a whole page still gets
its line back.

## v0.3.0

Nothing that compiled stops compiling. The native half now needs Ayra v0.9.1:
`go get github.com/arandu-io/ayra@v0.9.1`, and see Ayra's own `UPGRADE.md` for
what it changes in the app (sRGB blending, the 44 dp touch area, dialogs drawn
over the whole window).

What is drawn changes: the line is shorter (a mark's room is its figure and 4 dp
around it, and the area that answers a press is 44 dp, grown outward), and a
golden picture of an app that holds the line needs `-update`. To put the line on
a screen other than About, such as the foot of the sign-in screen, draw
`sig.LineProps{Enabled: true}.Layout(c, &state)`, with a `sig.State` of its own
for that place.

## v0.2.1

Nothing breaks: every new field defaults to v0.2.0's behaviour. To load the
stylesheet in `<head>`, add `{!! signature.Stylesheet() !!}` to the layout's
head and draw the line with `HeadStylesheet: true`.

## v0.2.0

Nothing breaks: a project with no licence behaves as on v0.1.0. To make a
partnership's signature permanent, add the licence HYZIS issued as a constant
in `bootstrap/app.go`, pass it with `AppURL` to `signature.New`, and append
`sig.Middleware()` to the application's middleware. In an Ayra app, pass the
same licence and the server's address in `sig.Options`.

## v0.1.0

The first release: nothing to upgrade from. Install it with
`go get github.com/hyz-is/arandu-signature@v0.1.0`, build it with
`signature.New` in `bootstrap/app.go`, and draw the line in the public footer.
