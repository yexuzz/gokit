package zaplog

import (
	"context"
	"fmt"

	"github.com/yexuzz/gokit/logx"
	"go.uber.org/zap"
)

// Logger 是基于 zap 的 logx.Logger 实现。
type Logger struct {
	zap *zap.Logger
}

var _ logx.Logger = (*Logger)(nil)

// NewDevelopment 创建适合本地终端观察的 zap 日志器。
func NewDevelopment(opts ...Option) (logx.Logger, error) {
	cfg := DefaultDevelopmentConfig()
	// 使用 Option 覆盖少量配置，避免调用方直接理解完整 Config。
	for _, opt := range opts {
		opt(&cfg)
	}
	return NewWithConfig(cfg)
}

// NewProduction 创建适合生产环境采集和落盘的 zap 日志器。
func NewProduction(opts ...Option) (logx.Logger, error) {
	cfg := DefaultProductionConfig()
	// 生产默认使用 JSON 输出，调用方只需要补充服务名、文件路径等少量信息。
	for _, opt := range opts {
		opt(&cfg)
	}
	return NewWithConfig(cfg)
}

// New 使用生产默认配置创建 zap 日志器。
func New(opts ...Option) (logx.Logger, error) {
	return NewProduction(opts...)
}

// NewWithConfig 使用完整配置创建 zap 日志器。
func NewWithConfig(cfg Config) (logx.Logger, error) {
	z, err := buildZapLogger(cfg)
	if err != nil {
		return nil, err
	}
	return &Logger{zap: z}, nil
}

// InitDevelopment 将 logx 全局默认日志器设置为开发模式 zap 日志器。
func InitDevelopment(opts ...Option) error {
	logger, err := NewDevelopment(opts...)
	if err != nil {
		return err
	}
	logx.Init(logger)
	return nil
}

// InitProduction 将 logx 全局默认日志器设置为生产模式 zap 日志器。
func InitProduction(opts ...Option) error {
	logger, err := NewProduction(opts...)
	if err != nil {
		return err
	}
	logx.Init(logger)
	return nil
}

// Sync 刷新 zap 日志缓冲，通常在进程退出前调用。
func (l *Logger) Sync() error {
	if l == nil || l.zap == nil {
		return nil
	}
	return l.zap.Sync()
}

// Debug 记录调试日志。
func (l *Logger) Debug(ctx context.Context, msg string, fields ...any) {
	l.log(ctx, logx.DebugLevel, msg, fields...)
}

// Info 记录普通业务日志。
func (l *Logger) Info(ctx context.Context, msg string, fields ...any) {
	l.log(ctx, logx.InfoLevel, msg, fields...)
}

// Warn 记录警告日志。
func (l *Logger) Warn(ctx context.Context, msg string, fields ...any) {
	l.log(ctx, logx.WarnLevel, msg, fields...)
}

// Error 记录错误日志。
func (l *Logger) Error(ctx context.Context, msg string, fields ...any) {
	l.log(ctx, logx.ErrorLevel, msg, fields...)
}

// log 写入一条 zap 结构化日志。
func (l *Logger) log(ctx context.Context, level logx.Level, msg string, fields ...any) {
	if l == nil || l.zap == nil {
		return
	}
	// 先追加上下文链路字段，再追加业务显式传入的字段。
	zapFields := normalizeLogFields(logx.ContextFields(ctx))
	zapFields = append(zapFields, normalizeFields(fields...)...)
	switch level {
	case logx.DebugLevel:
		l.zap.Debug(msg, zapFields...)
	case logx.WarnLevel:
		l.zap.Warn(msg, zapFields...)
	case logx.ErrorLevel:
		l.zap.Error(msg, zapFields...)
	default:
		l.zap.Info(msg, zapFields...)
	}
}

// normalizeLogFields 将 logx 字段列表转换为 zap 字段列表。
func normalizeLogFields(fields []logx.Field) []zap.Field {
	if len(fields) == 0 {
		return nil
	}
	zapFields := make([]zap.Field, 0, len(fields))
	for _, field := range fields {
		zapFields = append(zapFields, zap.Any(field.Key, field.Value))
	}
	return zapFields
}

// normalizeFields 将通用字段参数转换为 zap 字段。
func normalizeFields(fields ...any) []zap.Field {
	if len(fields) == 0 {
		return nil
	}
	zapFields := make([]zap.Field, 0, len(fields))
	for i := 0; i < len(fields); i++ {
		field := fields[i]
		switch v := field.(type) {
		case logx.Field:
			// typed field 直接转换，保留明确字段名和值。
			zapFields = append(zapFields, zap.Any(v.Key, v.Value))
		case error:
			zapFields = append(zapFields, zap.Error(v))
		case string:
			if i+1 < len(fields) {
				// 支持 "key", value 的宽松写法，方便框架和业务快速接入。
				zapFields = append(zapFields, zap.Any(v, fields[i+1]))
				i++
			} else {
				zapFields = append(zapFields, zap.String(fmt.Sprintf("field%d", i), v))
			}
		default:
			zapFields = append(zapFields, zap.Any(fmt.Sprintf("field%d", i), v))
		}
	}
	return zapFields
}
