package log

import (
	"flag"
	"log/slog"
	"os"
	"strings"
)

// Logger returns a logger with level parsed from args.
// Pass os.Args[1:] from main.
func Logger(args []string) *slog.Logger {
	level := parseLogLevelFlag(args)
	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: level,
	})
	return slog.New(handler)
}

func parseLogLevelFlag(args []string) *slog.LevelVar {
	var level slog.LevelVar
	level.Set(slog.LevelError) // default

	filtered := filterLogLevelArgs(args)
	if len(filtered) > 0 {
		fs := flag.NewFlagSet("early", flag.ExitOnError)
		fs.TextVar(&level, "log-level", &level, "")
		_ = fs.Parse(filtered)
	}

	return &level
}

// filterLogLevelArgs extracts only -log-level args from the slice.
// Handles both "-log-level=VALUE" and "-log-level VALUE" forms.
func filterLogLevelArgs(args []string) []string {
	for i, arg := range args {
		// Form: -log-level=VALUE or --log-level=VALUE
		if strings.HasPrefix(arg, "-log-level=") || strings.HasPrefix(arg, "--log-level=") {
			return []string{arg}
		}
		// Form: -log-level VALUE or --log-level VALUE
		if (arg == "-log-level" || arg == "--log-level") && i+1 < len(args) {
			return []string{arg, args[i+1]}
		}
	}
	return nil
}
