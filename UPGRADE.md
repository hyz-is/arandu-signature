# Upgrade Guide

## Unreleased

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
