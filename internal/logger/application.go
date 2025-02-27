package logger

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/lmittmann/tint"
)

var logger *slog.Logger

const LevelFatal slog.Level = slog.LevelError + 4

var (
	Debug = makeLogFunc(slog.LevelDebug)
	Info  = makeLogFunc(slog.LevelInfo)
	Warn  = makeLogFunc(slog.LevelWarn)
	Error = makeLogFunc(slog.LevelError)
	Fatal = makeLogFunc(LevelFatal)
)

func init() {
	w := os.Stderr
	logger = slog.New(
		tint.NewHandler(w, &tint.Options{
			Level:      slog.LevelDebug,
			TimeFormat: time.RFC3339,
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				if a.Key == slog.LevelKey {
					switch a.Value.Any() {
					case LevelFatal:
						return slog.Attr{
							Key:   slog.LevelKey,
							Value: slog.StringValue("\x1b[91mFATAL\x1b[0m"),
						}
					}
				}
				return a
			},
		}),
	)
	slog.SetDefault(logger)
}

func makeLogFunc(level slog.Level) func(msg string, args ...any) {
	return func(msg string, args ...any) {
		if len(args) > 0 {
			msg = fmt.Sprintf(msg, args...)
		}
		logger.Log(context.Background(), level, msg)

		if level == LevelFatal {
			os.Exit(1)
		}
	}
}
