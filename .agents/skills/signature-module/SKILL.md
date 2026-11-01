---
name: signature-module
description: Use when changing the signature's line, its assets, its stylesheet or the native half (Ayra widget, About screen, OpenURL).
---

# Changing the signature

- The web half is `kyse.go`: the marks under `marks/` are embedded and
  registered as view assets in `init`, the stylesheet is generated after them
  so its masks name the served addresses, and `Kyse(Props)` renders the line.
- The native half is `native/`: the marks are PNGs under `native/marks/`
  (rasterized from `marks/*.svg` at 96 px with a transparent background) and
  painted in the screen's ink; `OpenURL` has one file per platform
  (`open_android.go` by JNI, `open_js.go`, `open_ios.go`, `open_desktop.go`).
- After a change to the marks, rasterize the PNGs again from the SVGs.
- Gates: `go test -race ./...`, `go vet ./...`, `GOOS=js GOARCH=wasm go vet ./native/`
  and, with the NDK, `GOOS=android GOARCH=arm64 CGO_ENABLED=1 CC=<ndk clang> go vet ./native/`.
