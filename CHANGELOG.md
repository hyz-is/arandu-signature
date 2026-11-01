# Changelog

Everything worth knowing about a release of Arandu Signature is recorded here.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and
the versions follow [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

A published module version is immutable: Go serves it from the proxy forever, so
a release is corrected by another release and never by moving a tag.

## [Unreleased]

## [0.1.0] - 2026-11-01

### Added

- `signature.Kyse(Props)` and `signature.New(Config)`: the line "Desenvolvido
  por [HYZIS] em [Arandu]" for a Kyse layout, each mark a link to its site, in
  the surrounding ink, with its stylesheet and both marks registered as view
  assets. Off unless `Enabled`; Portuguese by default, English on request.
- `native.Line`, `native.About` and `native.OpenURL`: the same line for an Ayra
  app, with 40 dp mark targets, the About screen it belongs on, and opening the
  system browser on Android (ACTION_VIEW), in a browser build (js/wasm) and on
  the desktop. iOS draws the marks without a link for now.
