package zaplog

import (
	"github.com/yexuzz/gokit/logx"
	"github.com/yexuzz/gokit/logx/color"
)

// Option 表示 zap 日志器配置修改函数。
type Option func(cfg *Config)

// WithLevel 设置日志最小输出级别。
func WithLevel(level logx.Level) Option {
	return func(cfg *Config) {
		cfg.Level = level
	}
}

// WithConsole 设置是否写入终端。
func WithConsole(enabled bool) Option {
	return func(cfg *Config) {
		cfg.Console.Enabled = enabled
	}
}

// WithColor 设置终端日志级别是否使用颜色。
func WithColor(enabled bool) Option {
	return func(cfg *Config) {
		cfg.Console.Color = enabled
	}
}

// WithLevelColor 设置指定日志级别的终端颜色。
func WithLevelColor(level logx.Level, ansi color.ANSI) Option {
	return func(cfg *Config) {
		if cfg.Console.LevelColors == nil {
			cfg.Console.LevelColors = color.DefaultLevelColors()
		}
		cfg.Console.LevelColors[level] = ansi
	}
}

// WithLevelColors 批量设置日志级别的终端颜色。
func WithLevelColors(colors map[logx.Level]color.ANSI) Option {
	return func(cfg *Config) {
		if len(colors) == 0 {
			return
		}
		if cfg.Console.LevelColors == nil {
			cfg.Console.LevelColors = color.DefaultLevelColors()
		}
		for level, color := range colors {
			cfg.Console.LevelColors[level] = color
		}
	}
}

// WithFile 设置日志文件路径，并启用文件输出。
func WithFile(path string) Option {
	return func(cfg *Config) {
		cfg.File.Path = path
	}
}

// WithFileRotation 设置日志文件滚动策略。
func WithFileRotation(maxSize int, maxBackups int, maxAge int, compress bool) Option {
	return func(cfg *Config) {
		cfg.File.MaxSize = maxSize
		cfg.File.MaxBackups = maxBackups
		cfg.File.MaxAge = maxAge
		cfg.File.Compress = compress
	}
}

// WithCaller 设置是否记录调用位置。
func WithCaller(enabled bool) Option {
	return func(cfg *Config) {
		cfg.AddCaller = enabled
	}
}

// WithStacktrace 设置错误级别日志是否记录堆栈。
func WithStacktrace(enabled bool) Option {
	return func(cfg *Config) {
		cfg.AddStacktrace = enabled
	}
}

// WithServiceName 设置服务名称公共字段。
func WithServiceName(name string) Option {
	return func(cfg *Config) {
		cfg.ServiceName = name
	}
}
