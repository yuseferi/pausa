// Package version holds the application version. Override at build time
// with -ldflags "-X pausa/internal/version.Version=x.y.z". The default
// matches the latest released version.
package version

// Version is the current application version (without leading v).
var Version = "1.0.4"
