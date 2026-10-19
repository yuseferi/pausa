# Contributing to Pausa

Thanks for helping improve Pausa. This guide covers the local setup and the
conventions the project follows.

## Prerequisites

- macOS (Pausa is macOS-only at runtime; non-macOS builds exist only to keep CI
  cross-compiling)
- Go 1.25+
- Node.js 18+
- [Wails v2 CLI](https://wails.io/) — `go install github.com/wailsapp/wails/v2/cmd/wails@latest`
- Xcode Command Line Tools

## Getting started

```bash
git clone https://github.com/yuseferi/pausa.git
cd pausa

# The Go //go:embed directive needs frontend/dist to exist on a clean checkout.
mkdir -p frontend/dist && touch frontend/dist/.gitkeep

wails dev          # hot-reloading dev build
```

## Tests and checks

`make check` runs the core local checks:

```bash
make test    # go vet + go test -race + go build
make check   # the above, plus the Linux cross-compile, frontend lint + tests + build
make icons   # regenerate the app icon from build/icons/appicon.svg
```

CI runs the same steps **plus** two things `make check` does not: it generates
the Wails bindings (`wails generate module`) and runs `golangci-lint`. Run
`wails generate module` locally if you also need the generated bindings (the
frontend unit tests do not — they stub their boundaries).

Frontend:

```bash
cd frontend
npm run lint
npm run test        # vitest (unit + component), no generated bindings needed
npm run test:watch  # watch mode
npm run build
```

The Go scheduler test suite (`internal/scheduler`) is the most valuable part of
the project — it exercises every state transition with a fake clock and no
sleeps. Add tests there when you change scheduling behavior.

## Layout

See [`ARCHITECTURE.md`](ARCHITECTURE.md) for the concurrency model, state
machine, and where things live.

## Commits and releases

Pausa uses [Conventional Commits](https://www.conventionalcommits.org/) and
**semantic-release** on `main` — versions and changelogs are generated from
commit messages, so the prefix matters:

| Prefix | Release |
|---|---|
| `fix:` | patch |
| `feat:` | minor |
| `feat!:` / `BREAKING CHANGE:` | major |
| `docs:` `chore:` `test:` `ci:` `refactor:` | no release |

Keep commits focused. If a PR should not cut a release, use a non-`feat`/`fix`
prefix.

## Pull requests

- Sync with `origin/main` before you start, and keep the branch current.
- Fill in the PR template, including what you tested.
- Run `make check` (and `npm run test` in `frontend/`) before opening the PR.
- Runnable evidence beats prose: paste the command and its result.

## Accessibility

The UI should be usable with keyboard only and honour the system
**Reduce motion** setting. Icon-only controls need an `aria-label`; animated
components need a `prefers-reduced-motion` fallback.

## Reporting bugs and requesting features

Use the issue templates. For security issues, see [`SECURITY.md`](SECURITY.md).
