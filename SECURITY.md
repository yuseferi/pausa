# Security Policy

## Reporting a vulnerability

Please report security issues **privately** via GitHub's
[private vulnerability reporting](https://github.com/yuseferi/pausa/security/advisories/new)
rather than opening a public issue.

Include:

- what the issue is and its impact,
- steps to reproduce,
- the Pausa version and macOS version,
- any relevant logs (`~/Library/Logs/Pausa/pausa.log`).

You can expect an acknowledgement within a few days. Please give a reasonable
window to ship a fix before public disclosure.

## Scope

Pausa runs entirely on your Mac: no account, no telemetry, and no automatic
network requests (the update check runs only when you click it). Anything that
breaks those guarantees is in scope — for example:

- an unexpected network request or data leaving the machine,
- reading data beyond what the documented busy-detection signals require
  (microphone *in use*, Now Playing state, frontmost browser tab URL),
- privilege escalation or code execution through the app or its installer.

The build is currently unsigned and distributed via a Homebrew cask and a
one-line installer that clears the quarantine attribute. Tampering with the
release assets or the tap is in scope.

## Supported versions

Only the latest released version is supported with security fixes.
