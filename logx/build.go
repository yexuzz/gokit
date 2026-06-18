package logx

import (
	"strings"
)

// NewDevelopment 创建适合本地终端观察的 zap 日志器。
func NewDevelopment(opts ...Option) (*ZapLogger, error) {
	return Build(DevelopmentMode, opts...)
}

// NewProduction 创建适合生产环境采集和落盘的 zap 日志器。
func NewProduction(opts ...Option) (*ZapLogger, error) {
	return Build(ProductionMode, opts...)
}

// New 使用生产默认配置创建 zap 日志器。
func New(opts ...Option) (*ZapLogger, error) {
	return NewProduction(opts...)
}

// Build 按运行模式创建 zap 日志器，支持 dev/development 和 prod/production。
func Build(mode Mode, opts ...Option) (*ZapLogger, error) {
	cfg := DefaultConfig(mode)
	for _, opt := range opts {
		opt(&cfg)
	}
	return NewWithConfig(cfg)
}

// NewWithConfig 使用完整配置创建 zap 日志器。
func NewWithConfig(cfg Config) (*ZapLogger, error) {
	z, err := buildZapLogger(cfg)
	if err != nil {
		return nil, err
	}
	return &ZapLogger{zap: z}, nil
}

// InitDevelopment 将 logx 全局默认日志器设置为开发模式 zap 日志器。
func InitDevelopment(opts ...Option) error {
	return InitMode(DevelopmentMode, opts...)
}

// InitProduction 将 logx 全局默认日志器设置为生产模式 zap 日志器。
func InitProduction(opts ...Option) error {
	return InitMode(ProductionMode, opts...)
}

// InitMode 按运行模式创建 zap 日志器并写入 logx 全局默认日志器。
func InitMode(mode Mode, opts ...Option) error {
	logger, err := Build(mode, opts...)
	if err != nil {
		return err
	}
	Init(logger)
	return nil
}

// normalizeMode 归一化运行模式，未知模式按生产环境处理。
func normalizeMode(mode Mode) Mode {
	switch strings.ToLower(strings.TrimSpace(string(mode))) {
	case string(DevelopmentMode), "development":
		return DevelopmentMode
	default:
		return ProductionMode
	}
}
