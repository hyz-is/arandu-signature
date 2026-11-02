# Upgrade Guide

## Unreleased

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
