package logx

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

// ZapLogger 是基于 zap 的 logx.Logger 实现。
type ZapLogger struct {
	zap *zap.Logger
}

var _ Logger = (*ZapLogger)(nil)
var _ Syncer = (*ZapLogger)(nil)

// Sync 刷新 zap 日志缓冲，通常在进程退出前调用。
func (l *ZapLogger) Sync() error {
	if l == nil || l.zap == nil {
		return nil
	}
	return l.zap.Sync()
}

// Debug 记录调试日志。
func (l *ZapLogger) Debug(ctx context.Context, msg string, fields ...Field) {
	l.log(ctx, DebugLevel, msg, 0, fields...)
}

// Debugf 记录格式化调试日志。
func (l *ZapLogger) Debugf(ctx context.Context, format string, args ...any) {
	l.log(ctx, DebugLevel, fmt.Sprintf(format, args...), 0)
}

// Info 记录普通业务日志。
func (l *ZapLogger) Info(ctx context.Context, msg string, fields ...Field) {
	l.log(ctx, InfoLevel, msg, 0, fields...)
}

// Infof 记录格式化普通业务日志。
func (l *ZapLogger) Infof(ctx context.Context, format string, args ...any) {
	l.log(ctx, InfoLevel, fmt.Sprintf(format, args...), 0)
}

// Warn 记录警告日志。
func (l *ZapLogger) Warn(ctx context.Context, msg string, fields ...Field) {
	l.log(ctx, WarnLevel, msg, 0, fields...)
}

// Warnf 记录格式化警告日志。
func (l *ZapLogger) Warnf(ctx context.Context, format string, args ...any) {
	l.log(ctx, WarnLevel, fmt.Sprintf(format, args...), 0)
}

// Error 记录错误日志。
func (l *ZapLogger) Error(ctx context.Context, msg string, fields ...Field) {
	l.log(ctx, ErrorLevel, msg, 0, fields...)
}

// Errorf 记录格式化错误日志。
func (l *ZapLogger) Errorf(ctx context.Context, format string, args ...any) {
	l.log(ctx, ErrorLevel, fmt.Sprintf(format, args...), 0)
}

// log 写入一条 zap 结构化日志。
func (l *ZapLogger) log(ctx context.Context, level Level, msg string, callerSkip int, fields ...Field) {
	if l == nil || l.zap == nil {
		return
	}
	// 先追加上下文链路字段，再追加业务显式传入的字段。
	zapFields := normalizeLogFields(ContextFields(ctx))
	zapFields = append(zapFields, normalizeZapFields(fields...)...)
	logger := l.zap
	if callerSkip > 0 {
		logger = logger.WithOptions(zap.AddCallerSkip(callerSkip))
	}
	switch level {
	case DebugLevel:
		logger.Debug(msg, zapFields...)
	case WarnLevel:
		logger.Warn(msg, zapFields...)
	case ErrorLevel:
		logger.Error(msg, zapFields...)
	default:
		logger.Info(msg, zapFields...)
	}
}
