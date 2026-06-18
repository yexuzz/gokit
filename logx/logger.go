package logx

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/yexuzz/gokit/logx/color"
)

// Logger 定义 logx 对外提供的通用日志能力。
type Logger interface {
	// Debug 记录调试日志。
	Debug(ctx context.Context, msg string, fields ...Field)
	// Info 记录普通业务日志。
	Info(ctx context.Context, msg string, fields ...Field)
	// Warn 记录警告日志。
	Warn(ctx context.Context, msg string, fields ...Field)
	// Error 记录错误日志。
	Error(ctx context.Context, msg string, fields ...Field)
}

// Syncer 定义可选的日志缓冲刷新能力。
type Syncer interface {
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
		logFields := normalizeLooseFields(fields...)
		// 这里使用字符串级别是为了适配 ginx 这类不依赖 logx 类型的框架。
		switch Level(level) {
		case DebugLevel:
			logger.Debug(ctx, msg, logFields...)
		case WarnLevel:
			logger.Warn(ctx, msg, logFields...)
		case ErrorLevel:
			logger.Error(ctx, msg, logFields...)
		default:
			logger.Info(ctx, msg, logFields...)
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
func Debug(ctx context.Context, msg string, fields ...Field) {
	logger := Default()
	if zapLogger, ok := logger.(*ZapLogger); ok {
		zapLogger.log(ctx, DebugLevel, msg, 1, fields...)
		return
	}
	logger.Debug(ctx, msg, fields...)
}

// Info 使用全局默认日志器记录普通业务日志。
func Info(ctx context.Context, msg string, fields ...Field) {
	logger := Default()
	if zapLogger, ok := logger.(*ZapLogger); ok {
		zapLogger.log(ctx, InfoLevel, msg, 1, fields...)
		return
	}
	logger.Info(ctx, msg, fields...)
}

// Warn 使用全局默认日志器记录警告日志。
func Warn(ctx context.Context, msg string, fields ...Field) {
	logger := Default()
	if zapLogger, ok := logger.(*ZapLogger); ok {
		zapLogger.log(ctx, WarnLevel, msg, 1, fields...)
		return
	}
	logger.Warn(ctx, msg, fields...)
}

// Error 使用全局默认日志器记录错误日志。
func Error(ctx context.Context, msg string, fields ...Field) {
	logger := Default()
	if zapLogger, ok := logger.(*ZapLogger); ok {
		zapLogger.log(ctx, ErrorLevel, msg, 1, fields...)
		return
	}
	logger.Error(ctx, msg, fields...)
}

// Debugf 使用全局默认日志器记录格式化调试日志。
func Debugf(ctx context.Context, format string, args ...any) {
	Debug(ctx, fmt.Sprintf(format, args...))
}

// Infof 使用全局默认日志器记录格式化普通业务日志。
func Infof(ctx context.Context, format string, args ...any) {
	Info(ctx, fmt.Sprintf(format, args...))
}

// Warnf 使用全局默认日志器记录格式化警告日志。
func Warnf(ctx context.Context, format string, args ...any) {
	Warn(ctx, fmt.Sprintf(format, args...))
}

// Errorf 使用全局默认日志器记录格式化错误日志。
func Errorf(ctx context.Context, format string, args ...any) {
	Error(ctx, fmt.Sprintf(format, args...))
}

// Sync 刷新全局默认日志器的日志缓冲；默认日志器不支持刷新时直接返回 nil。
func Sync() error {
	syncer, ok := Default().(Syncer)
	if !ok {
		return nil
	}
	return syncer.Sync()
}

// Debug 记录默认终端调试日志。
func (stdLogger) Debug(ctx context.Context, msg string, fields ...Field) {
	writeStdLog(ctx, DebugLevel, msg, fields...)
}

// Info 记录默认终端普通日志。
func (stdLogger) Info(ctx context.Context, msg string, fields ...Field) {
	writeStdLog(ctx, InfoLevel, msg, fields...)
}

// Warn 记录默认终端警告日志。
func (stdLogger) Warn(ctx context.Context, msg string, fields ...Field) {
	writeStdLog(ctx, WarnLevel, msg, fields...)
}

// Error 记录默认终端错误日志。
func (stdLogger) Error(ctx context.Context, msg string, fields ...Field) {
	writeStdLog(ctx, ErrorLevel, msg, fields...)
}

// Debug 忽略空日志器调试日志。
func (nopLogger) Debug(ctx context.Context, msg string, fields ...Field) {}

// Info 忽略空日志器普通日志。
func (nopLogger) Info(ctx context.Context, msg string, fields ...Field) {}

// Warn 忽略空日志器警告日志。
func (nopLogger) Warn(ctx context.Context, msg string, fields ...Field) {}

// Error 忽略空日志器错误日志。
func (nopLogger) Error(ctx context.Context, msg string, fields ...Field) {}

// writeStdLog 写入根包默认终端日志。
func writeStdLog(ctx context.Context, level Level, msg string, fields ...Field) {
	allFields := ContextFields(ctx)
	lineColor, explicitFields := normalizeFields(fields...)
	allFields = append(allFields, explicitFields...)
	parts := make([]string, 0, len(allFields))
	for _, field := range allFields {
		parts = append(parts, fmt.Sprintf("%s=%v", field.Key, field.Value))
	}
	line := fmt.Sprintf("%s %-5s %s", time.Now().Format("2006-01-02 15:04:05.000"), strings.ToUpper(string(level)), msg)
	if len(parts) > 0 {
		line += " " + strings.Join(parts, " ")
	}
	if lineColor != color.None {
		line = string(lineColor) + line + string(color.Reset)
	}
	fmt.Fprintln(os.Stdout, line)
}

// normalizeFields 拆分终端颜色字段和真实结构化日志字段。
func normalizeFields(fields ...Field) (color.ANSI, []Field) {
	if len(fields) == 0 {
		return color.None, nil
	}
	lineColor := color.None
	logFields := make([]Field, 0, len(fields))
	for _, field := range fields {
		if field.Key == lineColorFieldKey {
			// 单条日志颜色只影响终端渲染，不作为结构化字段输出。
			lineColor = color.ANSI(fmt.Sprint(field.Value))
			continue
		}
		logFields = append(logFields, field)
	}
	return lineColor, logFields
}

// normalizeLooseFields 将框架适配层传入的宽松字段转换为显式 logx 字段。
func normalizeLooseFields(fields ...any) []Field {
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
				// 框架适配层保留 "key", value 写法，避免 ginx 之类的包依赖 logx.Field。
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
