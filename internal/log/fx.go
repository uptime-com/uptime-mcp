package log

import (
	"context"
	"log/slog"

	"go.uber.org/fx/fxevent"
)

// FxeventLogger returns an fxevent.Logger that uses slog.
// Regular fx lifecycle events are logged at DEBUG level (hidden by default),
// while errors are logged at ERROR level.
func FxeventLogger(ctx context.Context, logger *slog.Logger) fxevent.Logger {
	l := fxevent.SlogLogger{Logger: logger}
	l.UseContext(ctx)
	l.UseErrorLevel(slog.LevelError)
	l.UseLogLevel(slog.LevelDebug)
	return &l
}
