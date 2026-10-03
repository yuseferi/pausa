// Package version holds the application version. Release builds inject the
// real value at compile time with
// -ldflags "-X pausa/internal/version.Version=x.y.z"; the default marks a
// local/dev build.
package version

// Version is the current application version (without leading v).
var Version = "dev"
