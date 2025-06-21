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

func FatalLog(message string) {
	slog.Error(message)
	os.Exit(-1)
}

func Info(message string) {
	slog.Info(message)
}

func InitLogging() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, opts))
	slog.SetDefault(logger)
}
