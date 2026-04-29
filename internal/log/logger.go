// Package log configures slog to write to ~/Library/Logs/Pausa/pausa.log on
// macOS (and the platform equivalent elsewhere), with stderr fallback.
package log

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
)

// Init configures the default slog logger and returns the open log file
// (caller should Close on shutdown). If file logging fails for any reason,
// logging falls back to stderr only and the returned file is nil.
//
// The level can be overridden at runtime via the PAUSA_LOG_LEVEL env var:
// "debug", "info" (default), "warn", "error". Useful for diagnosing busy
// detection, scheduler transitions, etc. without recompiling.
func Init(defaultLevel slog.Level) *os.File {
	level := levelFromEnv(defaultLevel)

	path, err := logPath()
	var w io.Writer = os.Stderr
	var f *os.File
	if err == nil {
		if mkerr := os.MkdirAll(filepath.Dir(path), 0o755); mkerr == nil {
			file, oerr := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
			if oerr == nil {
				f = file
				w = io.MultiWriter(os.Stderr, file)
			}
		}
	}
	h := slog.NewTextHandler(w, &slog.HandlerOptions{Level: level})
	slog.SetDefault(slog.New(h))
	return f
}

func levelFromEnv(fallback slog.Level) slog.Level {
	switch os.Getenv("PAUSA_LOG_LEVEL") {
	case "debug", "DEBUG":
		return slog.LevelDebug
	case "info", "INFO":
		return slog.LevelInfo
	case "warn", "WARN", "warning", "WARNING":
		return slog.LevelWarn
	case "error", "ERROR":
		return slog.LevelError
	}
	return fallback
}

func logPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	// macOS convention; on other platforms this just lives under ~/Library
	// which is harmless for our cross-compile stubs.
	return filepath.Join(home, "Library", "Logs", "Pausa", "pausa.log"), nil
}
