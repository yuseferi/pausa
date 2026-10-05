## [1.3.1](https://github.com/yuseferi/pausa/compare/v1.3.0...v1.3.1) (2026-10-05)

### Bug Fixes

* **site:** use the explicit tap URL in Homebrew install commands ([#13](https://github.com/yuseferi/pausa/issues/13)) ([fe4af18](https://github.com/yuseferi/pausa/commit/fe4af18e884ca990c75dafece4fbfa78afe86245))

## [1.3.0](https://github.com/yuseferi/pausa/compare/v1.2.0...v1.3.0) (2026-10-03)

### Features

* add manual check-for-updates ([#11](https://github.com/yuseferi/pausa/issues/11)) ([06c2bb7](https://github.com/yuseferi/pausa/commit/06c2bb751111860811952f4884da0f8b43017149))

## [1.2.0](https://github.com/yuseferi/pausa/compare/v1.1.0...v1.2.0) (2026-10-03)

### Features

* add media-counts-as-activity idle toggle ([#12](https://github.com/yuseferi/pausa/issues/12)) ([ae9d2ae](https://github.com/yuseferi/pausa/commit/ae9d2aec1bc54598cfb20743f9b30a3d089c8aa1))

## [1.1.0](https://github.com/yuseferi/pausa/compare/v1.0.5...v1.1.0) (2026-10-03)

### Features

* add reset schedule button to dashboard and menu bar ([#9](https://github.com/yuseferi/pausa/issues/9)) ([4593c19](https://github.com/yuseferi/pausa/commit/4593c191c0a7d82256911fe6f89f61ae98ab4cb0))

### Bug Fixes

* don't count media playback as idle rest ([#10](https://github.com/yuseferi/pausa/issues/10)) ([2d897dd](https://github.com/yuseferi/pausa/commit/2d897dd4d0449ef1215fffcb708013f99cbb5545))

## [1.0.5](https://github.com/yuseferi/pausa/compare/v1.0.4...v1.0.5) (2026-10-02)

### Bug Fixes

* address PR review feedback ([8a97fc3](https://github.com/yuseferi/pausa/commit/8a97fc3fe19a3b403201be93b4ff0863580061f1))
* correct working-hours scheduling and config lifecycle bugs ([5f2eec5](https://github.com/yuseferi/pausa/commit/5f2eec5c296855c8b908d03c6d39c9ad3e57f178))
* harden break lifecycle and move busy sampling off the actor ([d4d5e72](https://github.com/yuseferi/pausa/commit/d4d5e7261563978c2e4b8e884a8b3e8d222497ec))
* make frontend resilient to empty config and stale listeners ([8108abe](https://github.com/yuseferi/pausa/commit/8108abea60fa4bc287003848f990650f47c0c7c6))

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
