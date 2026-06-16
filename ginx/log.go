package ginx

import (
	"context"
	"sync"
)

// LogFunc 是 ginx 对外暴露的日志事件出口。
type LogFunc func(ctx context.Context, level string, msg string, fields ...any)

var (
	logFuncMu sync.RWMutex
	logFunc   LogFunc = func(ctx context.Context, level string, msg string, fields ...any) {}
)

// SetLogFunc 设置 ginx 内部日志事件处理函数，传入 nil 时恢复为空实现。
func SetLogFunc(fn LogFunc) {
	logFuncMu.Lock()
	defer logFuncMu.Unlock()
	if fn == nil {
		// 恢复为空实现，保证未接入日志系统时 ginx 不产生额外副作用。
		logFunc = func(ctx context.Context, level string, msg string, fields ...any) {}
	} else {
		logFunc = fn
	}
}

// GetLogFunc 返回 ginx 当前使用的日志事件处理函数。
func GetLogFunc() LogFunc {
	logFuncMu.RLock()
	defer logFuncMu.RUnlock()
	return logFunc
}

// logDebug 写入 ginx 内部调试日志。
func logDebug(ctx context.Context, msg string, fields ...any) {
	GetLogFunc()(ctx, "debug", msg, fields...)
}

// logInfo 写入 ginx 内部普通日志。
func logInfo(ctx context.Context, msg string, fields ...any) {
	GetLogFunc()(ctx, "info", msg, fields...)
}

// logWarn 写入 ginx 内部警告日志。
func logWarn(ctx context.Context, msg string, fields ...any) {
	GetLogFunc()(ctx, "warn", msg, fields...)
}

// logError 写入 ginx 内部错误日志。
func logError(ctx context.Context, msg string, fields ...any) {
	GetLogFunc()(ctx, "error", msg, fields...)
}
