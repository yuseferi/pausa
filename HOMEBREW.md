# Homebrew Distribution

## Install

```bash
brew install --cask yuseferi/pausa/pausa
```

This auto-taps the public companion tap
[`yuseferi/homebrew-pausa`](https://github.com/yuseferi/homebrew-pausa) and
installs the latest release. Verified on Homebrew 7.x with
`brew install --cask --dry-run`: the fully-qualified name resolves and installs
without a separate `brew trust` step.

### Explicit-tap fallback

If the one-liner is unavailable (offline, older Homebrew, or you prefer the tap
from this repo):

```bash
brew tap yuseferi/pausa https://github.com/yuseferi/pausa
brew trust yuseferi/pausa   # Homebrew 6+ requires trusting third-party taps
brew install --cask pausa
```

Both cask copies select the correct asset with `on_arm` / `on_intel`, so the
install works on Apple Silicon (`arm64`) and Intel (`amd64`).

### First launch (unsigned builds)

Until the Apple secrets in [`NOTARIZATION.md`](NOTARIZATION.md) are configured,
released builds are **unsigned**, and recent Homebrew no longer offers
`--no-quarantine` (it removed the Gatekeeper bypass). macOS therefore blocks the
first launch. Users clear it once with:

```bash
xattr -dr com.apple.quarantine /Applications/pausa.app
```

or by right-clicking the app in Finder → **Open**. Configure the signing secrets
to ship notarized builds and remove this step.

---

## How the tap mirror works

Homebrew maps a fully-qualified `user/tap/name` to `github.com/user/homebrew-tap`
for auto-tapping. The app source lives in `yuseferi/pausa`, so a small companion
tap exists to satisfy that name.

- **Source of truth:** [`Casks/pausa.rb`](Casks/pausa.rb) in this repository.
- **Mirror:** `yuseferi/homebrew-pausa` (public) — only `Casks/pausa.rb` + a
  README. Do not edit it by hand; changes are overwritten.
- On each release, the `update-cask` job in
  [`.github/workflows/release.yml`](.github/workflows/release.yml):
  1. rewrites `Casks/pausa.rb` with the new version and SHA256s,
  2. commits it here,
  3. mirrors it to `yuseferi/homebrew-pausa`.

### Required secret: `TAP_GITHUB_TOKEN`

The mirror push needs a token that can write to the tap repository (the
workflow's `GITHUB_TOKEN` is scoped to this repo only).

1. Create a **fine-grained PAT** with **Contents: Read and write**, scoped to
   only `yuseferi/homebrew-pausa`.
2. Add it as a repository secret named **`TAP_GITHUB_TOKEN`**
   (Settings → Secrets and variables → Actions).

If the secret is absent, the mirror step prints
`::warning::TAP_GITHUB_TOKEN is not set; skipping the homebrew-pausa mirror.`
and the release still succeeds — the tap just stays at its last version.

---

## Release flow

1. **semantic-release** runs on every push to `main`. Conventional commits
   determine the next version; it creates the tag and GitHub release.
2. Because semantic-release publishes with `GITHUB_TOKEN`, and
   `GITHUB_TOKEN`-created events do not trigger other workflows, the
   `release: published` event never fires the build. The **Semantic Release
   workflow explicitly dispatches** `Build Release Assets` for the new tag
   (`workflow_dispatch` is exempt from that restriction). See
   `.github/workflows/semantic-release.yml`.
3. **Build Release Assets** (`release.yml`) builds the dual-arch zips, uploads
   them to the release, rewrites `Casks/pausa.rb`, and mirrors the tap.

Signing and notarization run when the Apple secrets are configured; otherwise
the workflow falls back to an unsigned build. See
[`NOTARIZATION.md`](NOTARIZATION.md).

---

## Local release prep

```bash
scripts/release.sh 1.0.2
# or
make release VERSION=1.0.2
```

This builds both zip assets and rewrites `Casks/pausa.rb`. Local runs do **not**
mirror the tap — use the CI release flow, or update
`yuseferi/homebrew-pausa` manually, when you need the tap to move.

To sign/notarize a local build, set the env vars from `NOTARIZATION.md` and run
`SIGN_NOTARIZE=1 scripts/release.sh <version>`.

---

## Current cask file

`Casks/pausa.rb` is kept close to Homebrew style:
- lower-case token
- `name`, `desc`, `homepage`
- `on_arm` / `on_intel` architecture-specific assets
- `app "pausa.app"`
- `zap` stanza

---

## Future: official Homebrew cask

Landing in `Homebrew/homebrew-cask` would remove the need for a tap entirely
(`brew install --cask pausa`). The remaining requirement in practice is a
signed and notarized app bundle — the plumbing now exists in `release.yml`
(see `NOTARIZATION.md`).

Suggested path:

1. Sign and notarize the app (configure the Apple secrets, cut a release).
2. Keep release assets stable and predictable (already dual-arch).
3. Copy `Casks/pausa.rb` into a branch of `Homebrew/homebrew-cask`.
4. Run `brew audit --cask pausa` from that repo.
5. Open a PR to `Homebrew/homebrew-cask`.
