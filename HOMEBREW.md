# Homebrew Distribution

## Current state

Pausa is already installable via a repo tap:

```bash
brew tap yuseferi/pausa https://github.com/yuseferi/pausa
brew install --cask pausa
```

That is the **current working path**.

---

## Goal: one-line install

To make this work:

```bash
brew install --cask pausa
```

the cask must be merged into **Homebrew/homebrew-cask**.

---

## Current blockers to upstream inclusion

### 1. App bundle is unsigned

At the moment:

```bash
codesign -dv --verbose=4 build/bin/pausa.app
```

reports that the app is **not signed**.

For an official Homebrew cask, signing is strongly preferred and in practice
often expected for a GUI app distributed as a `.app` bundle.

Recommended fix:
- Developer ID Application signing
- notarization with Apple
- staple the ticket to the final `.app`

### 2. Release is Apple Silicon only

The current release asset is:

```text
pausa-1.0.0-arm64-macos.zip
```

This means the cask is currently restricted to:

```ruby
depends_on arch: :arm64
```

This is valid, but a **universal** or dual-arch release is more likely to be
accepted and more useful to users.

Recommended fix:
- build and release both `darwin/arm64` and `darwin/amd64`
- or produce a universal app bundle if feasible

### 3. Ongoing cask maintenance

For each release, the cask must be updated with:
- new `version`
- new `sha256`
- release asset URL (if naming changes)

The current workflow automates building and uploading the release asset, but
does **not** yet auto-open a PR against `homebrew-cask`.

---

## Current cask file

The repo contains a cask candidate at:

```text
Casks/pausa.rb
```

It is intentionally shaped close to the Homebrew style already:
- lower-case token
- `name`, `desc`, `homepage`
- `depends_on arch: :arm64`
- `app "pausa.app"`
- `zap` stanza

---

## Suggested path to official inclusion

1. Sign and notarize the app
2. Produce either:
   - dual-arch releases, or
   - a universal build
3. Keep release assets stable and predictable
4. Copy `Casks/pausa.rb` into a branch of `Homebrew/homebrew-cask`
5. Run Homebrew audit from that repo
6. Open PR to `Homebrew/homebrew-cask`

---

## Example upstream cask PR draft

Title:

```text
Add pausa
```

Body:

```text
## Description

Adds Pausa, a native-feeling macOS break reminder for developers and
knowledge workers.

## App details

- homepage: https://github.com/yuseferi/pausa
- download: GitHub Releases
- package: signed / notarized .app bundle zip

## Notes

- overlays support fullscreen-app Spaces
- menu-bar app with native notifications
- busy-aware auto-pause for meetings and media
```

---

## If you only want the user-facing one-liner

The **only** durable way to get:

```bash
brew install --cask pausa
```

is to land in the official Homebrew cask repo.

Until then, the repo-tap install remains the realistic path.
