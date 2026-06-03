package logger

import (
	"io"
	"log/slog"
	"os"
)

// New creates and configures a new *slog.Logger.
// It dynamically selects either a JSON or Text handler based on the format parameter
// and sets the logging level based on the verbose parameter (Debug if true, otherwise Info).
func New(w io.Writer, verbose bool, format string) *slog.Logger {
	if w == nil {
		w = os.Stderr
	}

	level := slog.LevelInfo
	if verbose {
		level = slog.LevelDebug
	}

	opts := &slog.HandlerOptions{
		Level: level,
	}

	var handler slog.Handler
	if format == "json" {
		handler = slog.NewJSONHandler(w, opts)
	} else {
		handler = slog.NewTextHandler(w, opts)
	}

	return slog.New(handler)
}
