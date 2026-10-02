# Signing & notarization

Pausa's release workflow signs and notarizes macOS builds when the required
Apple secrets are configured. When they are absent (for example on forks or
PRs), the workflow falls back to an **unsigned** build and prints a warning, so
the pipeline keeps working.

## How it works

1. `wails build` produces an ad-hoc–signed `.app`.
2. [`scripts/sign-notarize.sh`](scripts/sign-notarize.sh) re-signs it with the
   Hardened Runtime and entitlements, submits it to Apple's notary service with
   `notarytool`, staples the ticket, verifies with `spctl`, and packages the
   `.zip` with `ditto` (required so the staple survives).
3. [`build/darwin/entitlements.plist`](build/darwin/entitlements.plist) grants
   `com.apple.security.automation.apple-events`, needed by the `osascript`
   browser detection. `Info.plist` carries `NSAppleEventsUsageDescription` for
   the one-time consent prompt.

A Wails app contains a single Mach-O executable and no nested frameworks, so one
`codesign` call on the bundle is sufficient. We never sign with `--deep`
(deprecated); it is only used for verification.

## Prerequisites

- A paid Apple Developer membership.
- A **Developer ID Application** certificate (not Apple Development / Mac App
  Distribution / ad-hoc).
- Either an App Store Connect **API key** (`.p8`) or an **Apple ID app-specific
  password**.

## Repository secrets

Add these under **Settings → Secrets and variables → Actions**.

| Secret | Required | Notes |
|---|---|---|
| `APPLE_CERTIFICATE_P12_BASE64` | yes | `base64 -i DeveloperID.p12` |
| `APPLE_CERTIFICATE_PASSWORD` | yes | the password used when exporting the `.p12` |
| `APPLE_SIGNING_IDENTITY` | no | e.g. `Developer ID Application: Your Name (TEAMID)`; auto-detected if omitted |
| `NOTARY_KEY_P8_BASE64` | one of | `base64 -i AuthKey_XXXXXXXXXX.p8` |
| `NOTARY_KEY_ID` | with key | the API key's ID |
| `NOTARY_ISSUER` | team keys | the Issuer ID; **omit** for individual keys |
| `APPLE_ID` | one of | alternative to the API key |
| `APPLE_APP_PASSWORD` | with Apple ID | app-specific password |
| `APPLE_TEAM_ID` | with Apple ID | your 10-character Team ID |

Provide **either** the API-key group or the Apple-ID group, not both.

## Creating the certificate (.p12)

1. Keychain Access → **Certificate Assistant → Request a Certificate from a
   Certificate Authority** → save the CSR to disk.
2. [developer.apple.com](https://developer.apple.com/account/resources/certificates)
   → Certificates → **+** → **Developer ID Application**.
3. Upload the CSR and download the `.cer`, then double-click it to add it to your
   keychain.
4. In Keychain Access, locate **Developer ID Application: …** under
   *My Certificates*, right-click → **Export** → `.p12` with a password.
5. `base64 -i DeveloperID.p12 | pbcopy` and paste into
   `APPLE_CERTIFICATE_P12_BASE64`.

## Creating the API key

1. [appstoreconnect.apple.com](https://appstoreconnect.apple.com) → Users and
   Access → Integrations → **App Store Connect API** → Team Keys → **+**
   (role: Developer).
2. Download `AuthKey_XXXXXXXXXX.p8` (only downloadable once) and note the
   **Key ID** and **Issuer ID**.
3. `base64 -i AuthKey_XXXXXXXXXX.p8 | pbcopy` → `NOTARY_KEY_P8_BASE64`; set
   `NOTARY_KEY_ID` and `NOTARY_ISSUER`.

## Testing locally

With your Developer ID certificate in the login keychain:

```bash
SIGN_NOTARIZE=1 \
NOTARY_KEY_PATH="$HOME/keys/AuthKey_XXXXXXXXXX.p8" \
NOTARY_KEY_ID="XXXXXXXXXX" \
NOTARY_ISSUER="xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx" \
scripts/release.sh 1.0.6
```

Or sign/notarize an existing bundle directly:

```bash
scripts/sign-notarize.sh build/bin/pausa.app dist/pausa.app.zip
```

## Verifying a downloaded build

```bash
spctl --assess --type install -vvv /Applications/pausa.app   # → accepted, source=Notarized Developer ID
codesign -dvv /Applications/pausa.app                        # expect Timestamp= and flags=…(runtime)
xcrun stapler validate /Applications/pausa.app
```

## Notes

- Notarization is per build; every release must be notarized again.
- The first time Pausa reads a browser tab, macOS shows an Automation prompt
  (System Settings → Privacy & Security → Automation). This cannot be exercised
  in CI — the check simply no-ops there.
- Never add `com.apple.security.get-task-allow` to the entitlements; the notary
  service rejects it.
- The notary service accepts a `.zip` but cannot staple it — we staple the
  `.app` and re-zip.
