package logx

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

// Logger 定义 logx 对外提供的通用日志能力。
type Logger interface {
	// Debug 记录调试日志，通常用于开发环境排查细节。
	Debug(ctx context.Context, msg string, fields ...any)
	// Info 记录普通业务日志，通常用于描述关键流程节点。
	Info(ctx context.Context, msg string, fields ...any)
	// Warn 记录警告日志，通常用于描述可恢复的异常状态。
	Warn(ctx context.Context, msg string, fields ...any)
	// Error 记录错误日志，通常用于描述请求处理失败或系统异常。
	Error(ctx context.Context, msg string, fields ...any)
	// Sync 刷新日志缓冲，通常在进程退出前调用。
	Sync() error
}

// stdLogger 是 logx 根包提供的轻量默认日志器。
type stdLogger struct{}

// nopLogger 是不会输出任何内容的空日志器。
type nopLogger struct{}

var (
	defaultMu     sync.RWMutex
	defaultLogger Logger = stdLogger{}
)

// NewNop 创建一个不会写入任何日志的空日志器。
func NewNop() Logger {
	return nopLogger{}
}

// Init 写入全局默认日志器，传入 nil 时恢复为空日志器。
func Init(logger Logger) {
	defaultMu.Lock()
	defer defaultMu.Unlock()
	if logger == nil {
		// nil 表示主动关闭默认日志输出，方便测试或特殊场景静默运行。
		defaultLogger = NewNop()
		return
	}
	defaultLogger = logger
}

// Default 返回全局默认日志器。
func Default() Logger {
	defaultMu.RLock()
	defer defaultMu.RUnlock()
	return defaultLogger
}

// AsLogFunc 将 logx 日志器转换为框架可注入的函数出口。
func AsLogFunc(logger Logger) func(ctx context.Context, level string, msg string, fields ...any) {
	return func(ctx context.Context, level string, msg string, fields ...any) {
		if logger == nil {
			return
		}
		// 这里使用字符串级别是为了适配 ginx 这类不依赖 logx 类型的框架。
		switch Level(level) {
		case DebugLevel:
			logger.Debug(ctx, msg, fields...)
		case WarnLevel:
			logger.Warn(ctx, msg, fields...)
		case ErrorLevel:
			logger.Error(ctx, msg, fields...)
		default:
			logger.Info(ctx, msg, fields...)
		}
	}
}

// DefaultLogFunc 返回基于全局默认日志器的函数式日志出口。
func DefaultLogFunc() func(ctx context.Context, level string, msg string, fields ...any) {
	return func(ctx context.Context, level string, msg string, fields ...any) {
		// 每次调用都读取当前默认日志器，确保 Init 后接入方自动使用新配置。
		AsLogFunc(Default())(ctx, level, msg, fields...)
	}
}

// Debug 使用全局默认日志器记录调试日志。
func Debug(ctx context.Context, msg string, fields ...any) {
	Default().Debug(ctx, msg, fields...)
}

// Info 使用全局默认日志器记录普通业务日志。
func Info(ctx context.Context, msg string, fields ...any) {
	Default().Info(ctx, msg, fields...)
}

// Warn 使用全局默认日志器记录警告日志。
func Warn(ctx context.Context, msg string, fields ...any) {
	Default().Warn(ctx, msg, fields...)
}

// Error 使用全局默认日志器记录错误日志。
func Error(ctx context.Context, msg string, fields ...any) {
	Default().Error(ctx, msg, fields...)
}

// Sync 使用全局默认日志器刷新日志缓冲。
func Sync() error {
	return Default().Sync()
}

// Debug 记录默认终端调试日志。
func (stdLogger) Debug(ctx context.Context, msg string, fields ...any) {
	writeStdLog(ctx, DebugLevel, msg, fields...)
}

// Info 记录默认终端普通日志。
func (stdLogger) Info(ctx context.Context, msg string, fields ...any) {
	writeStdLog(ctx, InfoLevel, msg, fields...)
}

// Warn 记录默认终端警告日志。
func (stdLogger) Warn(ctx context.Context, msg string, fields ...any) {
	writeStdLog(ctx, WarnLevel, msg, fields...)
}

// Error 记录默认终端错误日志。
func (stdLogger) Error(ctx context.Context, msg string, fields ...any) {
	writeStdLog(ctx, ErrorLevel, msg, fields...)
}

// Sync 刷新默认终端日志缓冲。
func (stdLogger) Sync() error {
	return nil
}

// Debug 忽略空日志器调试日志。
func (nopLogger) Debug(ctx context.Context, msg string, fields ...any) {}

// Info 忽略空日志器普通日志。
func (nopLogger) Info(ctx context.Context, msg string, fields ...any) {}

// Warn 忽略空日志器警告日志。
func (nopLogger) Warn(ctx context.Context, msg string, fields ...any) {}

// Error 忽略空日志器错误日志。
func (nopLogger) Error(ctx context.Context, msg string, fields ...any) {}

// Sync 刷新空日志器缓冲。
func (nopLogger) Sync() error {
	return nil
}

// writeStdLog 写入根包默认终端日志。
func writeStdLog(ctx context.Context, level Level, msg string, fields ...any) {
	allFields := ContextFields(ctx)
	allFields = append(allFields, normalizeFields(fields...)...)
	parts := make([]string, 0, len(allFields))
	for _, field := range allFields {
		parts = append(parts, fmt.Sprintf("%s=%v", field.Key, field.Value))
	}
	line := fmt.Sprintf("%s %-5s %s", time.Now().Format("2006-01-02 15:04:05.000"), strings.ToUpper(string(level)), msg)
	if len(parts) > 0 {
		line += " " + strings.Join(parts, " ")
	}
	fmt.Fprintln(os.Stdout, line)
}

// normalizeFields 将通用字段参数转换为 logx 字段。
func normalizeFields(fields ...any) []Field {
	if len(fields) == 0 {
		return nil
	}
	logFields := make([]Field, 0, len(fields))
	for i := 0; i < len(fields); i++ {
		field := fields[i]
		switch v := field.(type) {
		case Field:
			logFields = append(logFields, v)
		case error:
			logFields = append(logFields, Err(v))
		case string:
			if i+1 < len(fields) {
				// 支持 "key", value 的宽松写法，方便框架和业务快速接入。
				logFields = append(logFields, Any(v, fields[i+1]))
				i++
			} else {
				logFields = append(logFields, String(fmt.Sprintf("field%d", i), v))
			}
		default:
			logFields = append(logFields, Any(fmt.Sprintf("field%d", i), v))
		}
	}
	return logFields
}
