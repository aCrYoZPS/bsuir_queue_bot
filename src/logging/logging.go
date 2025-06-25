package logging

import (
	"log/slog"
	"os"
	"time"
)

var opts = &slog.HandlerOptions{
	AddSource: false,
	Level:     slog.LevelDebug,
	ReplaceAttr: func(groups []string, attr slog.Attr) slog.Attr {
		if attr.Key != slog.TimeKey {
			return attr
		}

		curTime := attr.Value.Time()

		attr.Value = slog.StringValue(curTime.Format(time.DateTime))
		return attr
	},
}

func FatalLog(message string, args ...any) {
	slog.Error(message, args...)
	os.Exit(-1)
}

func Info(message string, args ...any) {
	slog.Info(message, args...)
}

func Error(message string, args ...any) {
	slog.Error(message, args...)
}

func InitLogging() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, opts))
	slog.SetDefault(logger)
}
