# Changelog

Everything worth knowing about a release of Arandu Signature is recorded here.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and
the versions follow [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

A published module version is immutable: Go serves it from the proxy forever, so
a release is corrected by another release and never by moving a tag.

## [Unreleased]

## [0.2.1] - 2026-11-01

### Added

- `Stylesheet()`, the line's `<link>` for a layout's `<head>`, and
  `Props.HeadStylesheet`, which draws the line without carrying its own link
  into the body. On a licensed project the middleware puts the link back in
  `<head>` when a page still signs but lost it.
- `Props.Size` (`SizeSmall`, `SizeLarge`) and the `--arandu-signature-scale`
  custom property for the marks' height on the web; `Options.Height` in the
  native package, the Arandu mark keeping its 17-to-20 proportion.
- `AboutProps.Back` and `AboutProps.Eyebrow`, for the app's own way back and
  small title, and `AboutProps.Placement` (`AtFoot`, the default, or
  `AfterContent`, centred right after the description). `Name` is optional
  when `Mark` already carries the wordmark.
- `OpenURL` opens the address through UIKit on iOS.

## [0.2.0] - 2026-11-01

### Added

- Partnership licences: `Config.License` and `Config.AppURL`, and
  `Options.License` and `Options.Host` in the native package. A licence HYZIS
  signs for the project's domain and term forces the signature on, ignoring
  `Enabled`; an invalid or expired one is logged and ignored. `Verify` reads
  one; `localhost` and loopback addresses match any licence.
- `Module.Middleware()`: on a licensed project, puts the line back before
  `</body>` on any whole HTML page that lacks its `data-arandu-signature`
  marker.
- `cmd/signature-license`: generates HYZIS's key (`-genkey`) and issues
  licences.

### Changed

- The line carries the `data-arandu-signature` marker, and its stylesheet
  keeps it visible against common hiding rules.

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
