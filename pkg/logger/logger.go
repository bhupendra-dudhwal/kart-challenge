package logger

import (
	"context"
	"fmt"
	"kart-challenge/internal/constants"
	"kart-challenge/internal/core/ports"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type zapLogger struct {
	logger *zap.Logger
	level  zap.AtomicLevel
}

func NewLogger(logeLevel constants.LogLevel, env constants.Environment) (ports.LoggerPorts, error) {
	var cfg zap.Config

	if env != constants.ENV_PRODUCTION {
		cfg = zap.NewDevelopmentConfig()
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		cfg = zap.NewProductionConfig()
		cfg.EncoderConfig.TimeKey = "time"
	}
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	level := zapcore.InfoLevel
	if err := level.UnmarshalText([]byte(logeLevel)); err != nil {
		fmt.Printf("invalid log level '%s', defaulting to INFO\n", logeLevel)
	}

	atomicLevel := zap.NewAtomicLevelAt(level)
	cfg.Level = atomicLevel

	// Build logger
	coreLogger, err := cfg.Build(
		zap.AddCaller(),
		zap.AddCallerSkip(2),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)
	if err != nil {
		return nil, fmt.Errorf("zap logger init failed: %w", err)
	}

	return &zapLogger{
		logger: coreLogger,
		level:  atomicLevel,
	}, nil
}

func (l *zapLogger) log(ctx context.Context, level zapcore.Level, msg string, fields ...zap.Field) {
	if ce := l.logger.Check(level, msg); ce != nil {
		ce.Write(append(zapFieldsFromContext(ctx), fields...)...)
	}
}

// With
func (l *zapLogger) With(fields ...zap.Field) ports.LoggerPorts {
	return &zapLogger{logger: l.logger.With(fields...), level: l.level}
}

// Without context
func (l *zapLogger) Info(msg string, fields ...zap.Field) {
	l.log(nil, zapcore.InfoLevel, msg, fields...)
}

func (l *zapLogger) Error(msg string, fields ...zap.Field) {
	l.log(nil, zapcore.ErrorLevel, msg, fields...)
}

func (l *zapLogger) Warn(msg string, fields ...zap.Field) {
	l.log(nil, zapcore.WarnLevel, msg, fields...)
}

func (l *zapLogger) Debug(msg string, fields ...zap.Field) {
	l.log(nil, zapcore.DebugLevel, msg, fields...)
}

// With context
func (l *zapLogger) InfoCtx(ctx context.Context, msg string, fields ...zap.Field) {
	l.log(ctx, zapcore.InfoLevel, msg, fields...)
}

func (l *zapLogger) ErrorCtx(ctx context.Context, msg string, fields ...zap.Field) {
	l.log(ctx, zapcore.ErrorLevel, msg, fields...)
}

func (l *zapLogger) WarnCtx(ctx context.Context, msg string, fields ...zap.Field) {
	l.log(ctx, zapcore.WarnLevel, msg, fields...)
}

func (l *zapLogger) DebugCtx(ctx context.Context, msg string, fields ...zap.Field) {
	l.log(ctx, zapcore.DebugLevel, msg, fields...)
}

func zapFieldsFromContext(ctx context.Context) []zap.Field {
	if ctx == nil {
		return nil
	}

	fields := make([]zap.Field, 0, 2)

	if reqID, ok := ctx.Value(constants.CtxRequestID).(string); ok {
		fields = append(fields, zap.String(constants.CtxRequestID.String(), reqID))
	}

	if traceID, ok := ctx.Value(constants.CtxTraceID).(string); ok {
		fields = append(fields, zap.String(constants.CtxTraceID.String(), traceID))
	}

	return fields
}

func (l *zapLogger) SetLevel(level zapcore.Level) {
	l.level.SetLevel(level)
}

func (l *zapLogger) Close() error {
	return l.logger.Sync()
}
