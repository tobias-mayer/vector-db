package log

import (
	"context"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

type Logger interface {
	With(ctx context.Context, args ...interface{}) Logger

	Debug(args ...interface{})
	Info(args ...interface{})
	Error(args ...interface{})
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Errorf(format string, args ...interface{})
}

type logger struct {
	*zap.SugaredLogger
}

type contextKey int

const (
	requestIdKey contextKey = iota
	correlationIdKey
)

func New() Logger {
	l, _ := zap.NewProduction()
	return NewWithZap(l)
}

func NewWithZap(l *zap.Logger) Logger {
	return &logger{l.Sugar()}
}

func NewForTest() (Logger, *observer.ObservedLogs) {
	core, recorded := observer.New(zapcore.InfoLevel)
	return NewWithZap(zap.New(core)), recorded
}

func (l *logger) With(ctx context.Context, args ...interface{}) Logger {
	if ctx != nil {
		if id, ok := ctx.Value(requestIdKey).(string); ok {
			args = append(args, zap.String("request_id", id))
		}
		if id, ok := ctx.Value(correlationIdKey).(string); ok {
			args = append(args, zap.String("correlation_id", id))
		}
	}
	if len(args) > 0 {
		return &logger{l.SugaredLogger.With(args...)}
	}
	return l
}
