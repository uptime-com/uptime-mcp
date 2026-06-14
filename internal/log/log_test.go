package log

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogger(t *testing.T) {
	logger := Logger(nil)
	assert.NotNil(t, logger)
}

func TestParseLogLevelFlag(t *testing.T) {
	t.Run("default is ERROR", func(t *testing.T) {
		level := parseLogLevelFlag(nil)
		assert.Equal(t, slog.LevelError, level.Level())
	})

	t.Run("parses -log-level=DEBUG", func(t *testing.T) {
		level := parseLogLevelFlag([]string{"-log-level=DEBUG"})
		assert.Equal(t, slog.LevelDebug, level.Level())
	})

	t.Run("parses -log-level INFO", func(t *testing.T) {
		level := parseLogLevelFlag([]string{"-log-level", "INFO"})
		assert.Equal(t, slog.LevelInfo, level.Level())
	})

	t.Run("parses --log-level=WARN", func(t *testing.T) {
		level := parseLogLevelFlag([]string{"--log-level=WARN"})
		assert.Equal(t, slog.LevelWarn, level.Level())
	})

	t.Run("finds log-level after unknown flags", func(t *testing.T) {
		level := parseLogLevelFlag([]string{"-unknown", "-log-level=WARN"})
		assert.Equal(t, slog.LevelWarn, level.Level())
	})

	t.Run("finds log-level between other flags", func(t *testing.T) {
		level := parseLogLevelFlag([]string{"-foo", "-log-level", "DEBUG", "-bar"})
		assert.Equal(t, slog.LevelDebug, level.Level())
	})
}

func TestFilterLogLevelArgs(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want []string
	}{
		{"nil args", nil, nil},
		{"empty args", []string{}, nil},
		{"no log-level", []string{"-foo", "-bar"}, nil},
		{"-log-level=VALUE", []string{"-log-level=DEBUG"}, []string{"-log-level=DEBUG"}},
		{"--log-level=VALUE", []string{"--log-level=INFO"}, []string{"--log-level=INFO"}},
		{"-log-level VALUE", []string{"-log-level", "WARN"}, []string{"-log-level", "WARN"}},
		{"--log-level VALUE", []string{"--log-level", "ERROR"}, []string{"--log-level", "ERROR"}},
		{"after other flags", []string{"-x", "-log-level=DEBUG"}, []string{"-log-level=DEBUG"}},
		{"between flags", []string{"-a", "-log-level", "INFO", "-b"}, []string{"-log-level", "INFO"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filterLogLevelArgs(tt.args)
			assert.Equal(t, tt.want, got)
		})
	}
}
