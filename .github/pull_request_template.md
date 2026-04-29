## Summary

- 

## Testing

- [ ] `go build ./...`
- [ ] `go test -race ./...`
- [ ] `npm run build` (in `frontend/`)
- [ ] manual macOS verification if native behavior changed

## Release impact

Use **Conventional Commits** in the squash/rebase commit message so semantic-release can version correctly.

Examples:

- `fix: pause scheduler during muted frontmost YouTube playback`
- `feat: add Firefox browser media detection`
- `docs: clarify Homebrew unsigned-install flow`

Versioning rules:

- `fix:` -> patch release
- `feat:` -> minor release
- `feat!:` or `BREAKING CHANGE:` -> major release
