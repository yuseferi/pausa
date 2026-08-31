## [1.0.4](https://github.com/yuseferi/pausa/compare/v1.0.3...v1.0.4) (2026-08-31)

### Bug Fixes

* wire up dormant settings and fix scheduler pause bugs ([5302779](https://github.com/yuseferi/pausa/commit/5302779c53003c2f0620360d79190161e6b28c85)), closes [#3](https://github.com/yuseferi/pausa/issues/3)

## [1.0.3](https://github.com/yuseferi/pausa/compare/v1.0.2...v1.0.3) (2026-08-31)

### Bug Fixes

* **cask:** update arm64 sha256 for v1.0.1 ([6d4da18](https://github.com/yuseferi/pausa/commit/6d4da1814d6c02bbb9ab4a5f7cfb67237834184f))

## [1.0.2](https://github.com/yuseferi/pausa/compare/v1.0.1...v1.0.2) (2026-04-30)

### Bug Fixes

* resolve golangci-lint errors (errcheck, staticcheck, unused) ([aa40835](https://github.com/yuseferi/pausa/commit/aa40835b236dddc1c132153715b4beafaddb423f))

# Changelog

All notable changes to this project will be documented in this file.

The format is based on semantic-release generated notes.

## [1.0.1](https://github.com/yuseferi/pausa/releases/tag/v1.0.1) - 2026-04-29

- Added dual-architecture Homebrew tap distribution for macOS (`arm64` and `amd64`)
- Added native fullscreen-space overlays across monitors
- Added busy-aware auto-pause for meetings and media playback
- Added browser/video heuristics for muted frontmost media pages
- Added release tooling (`make release`, local release helper, release workflows)
