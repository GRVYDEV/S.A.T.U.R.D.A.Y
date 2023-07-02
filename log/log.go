package log

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"golang.org/x/exp/slog"
)

var level = new(slog.LevelVar)

func SetLevel(lv slog.Level) {
	level.Set(lv)
}

type Logger struct {
	*slog.Logger
}

func New() Logger {
	replace := func(groups []string, a slog.Attr) slog.Attr {
		if a.Key == slog.SourceKey {
			source := a.Value.Any().(*slog.Source)
			source.File = filepath.Base(source.File)
		}

		return a
	}
	return Logger{
		slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			AddSource:   true,
			ReplaceAttr: replace,
			Level:       level,
		})),
	}
}

func (l Logger) logf(level slog.Level, format string, v ...any) {
	if !l.Enabled(context.Background(), level) {
		return
	}

	var pcs [1]uintptr
	runtime.Callers(3, pcs[:])
	r := slog.NewRecord(time.Now(), level, fmt.Sprintf(format, v...), pcs[0])
	_ = l.Handler().Handle(context.Background(), r)
}

func (l Logger) log(level slog.Level, msg string, args ...any) {
	if !l.Enabled(context.Background(), level) {
		return
	}

	var pcs [1]uintptr
	runtime.Callers(3, pcs[:])
	r := slog.NewRecord(time.Now(), level, msg, pcs[0])
	r.Add(args...)
	_ = l.Handler().Handle(context.Background(), r)
}

func (l Logger) Debugf(format string, v ...any) {
	l.logf(slog.LevelDebug, format, v...)
}

func (l Logger) Infof(format string, v ...any) {
	l.logf(slog.LevelInfo, format, v...)
}

func (l Logger) Warnf(format string, v ...any) {
	l.logf(slog.LevelWarn, format, v...)
}

func (l Logger) Error(err error, msg string, args ...any) {
	args = append(args, "error")
	args = append(args, err)
	l.log(slog.LevelError, msg, args...)
}

func (l Logger) Errorf(err error, format string, v ...any) {
	l.log(slog.LevelError, fmt.Sprintf(format, v...), err)
}

func (l Logger) Fatal(err error, msg string, args ...any) {
	args = append(args, "error")
	args = append(args, err)
	l.log(slog.LevelError, msg, args...)
	os.Exit(1)
}

func (l Logger) Fatalf(err error, format string, v ...any) {
	l.log(slog.LevelError, fmt.Sprintf(format, v...), err)
	os.Exit(1)
}
