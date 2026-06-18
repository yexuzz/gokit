package logx

import (
	"github.com/yexuzz/gokit/logx/color"
)

// Option 表示 zap 日志器配置修改函数。
type Option func(cfg *Config)

// FileOption 表示文件日志配置修改函数。
type FileOption func(cfg *FileConfig)

// WithLevel 设置日志最小输出级别。
func WithLevel(level Level) Option {
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

// WithLineColor 设置是否按日志级别给整行终端日志着色。
func WithLineColor(enabled bool) Option {
	return func(cfg *Config) {
		cfg.Console.LineColor = enabled
	}
}

// WithLevelColor 设置指定日志级别的终端颜色。
func WithLevelColor(level Level, ansi color.ANSI) Option {
	return func(cfg *Config) {
		if cfg.Console.LevelColors == nil {
			cfg.Console.LevelColors = defaultLevelColors()
		}
		cfg.Console.LevelColors[level] = ansi
	}
}

// WithLevelColors 批量设置日志级别的终端颜色。
func WithLevelColors(colors map[Level]color.ANSI) Option {
	return func(cfg *Config) {
		if len(colors) == 0 {
			return
		}
		if cfg.Console.LevelColors == nil {
			cfg.Console.LevelColors = defaultLevelColors()
		}
		for level, color := range colors {
			cfg.Console.LevelColors[level] = color
		}
	}
}

// WithOnlyLevelColor 设置只有指定日志级别显示颜色，其它级别不显示颜色。
func WithOnlyLevelColor(level Level, ansi color.ANSI) Option {
	return func(cfg *Config) {
		cfg.Console.LevelColors = noneLevelColors()
		cfg.Console.LevelColors[level] = ansi
	}
}

// WithFile 启用按日期切割的文件输出；未传配置时使用默认日志目录和命名规则。
func WithFile(opts ...FileOption) Option {
	return func(cfg *Config) {
		cfg.File.Enabled = true
		cfg.File = normalizeFileConfig(cfg.File)
		for _, opt := range opts {
			if opt != nil {
				opt(&cfg.File)
			}
		}
		cfg.File = normalizeFileConfig(cfg.File)
	}
}

// WithoutFile 关闭文件日志输出，只保留其它已启用的输出方式。
func WithoutFile() Option {
	return func(cfg *Config) {
		cfg.File.Enabled = false
	}
}

// FileDir 设置日志文件根目录。
func FileDir(dir string) FileOption {
	return func(cfg *FileConfig) {
		cfg.Dir = dir
	}
}

// FileDateLayout 设置日志日期目录格式。
func FileDateLayout(layout string) FileOption {
	return func(cfg *FileConfig) {
		cfg.DateLayout = layout
	}
}

// FileExt 设置日志文件扩展名，传入 log 时会自动补全为 .log。
func FileExt(ext string) FileOption {
	return func(cfg *FileConfig) {
		cfg.Ext = ext
	}
}

// FileInfoName 设置普通日志文件名，不需要包含扩展名。
func FileInfoName(name string) FileOption {
	return FileInfo(name)
}

// FileInfo 设置普通日志文件名，不需要包含扩展名。
func FileInfo(name string) FileOption {
	return func(cfg *FileConfig) {
		cfg.InfoFileName = name
	}
}

// FileErrorName 设置错误日志文件名，不需要包含扩展名。
func FileErrorName(name string) FileOption {
	return FileError(name)
}

// FileError 设置错误日志文件名，不需要包含扩展名。
func FileError(name string) FileOption {
	return func(cfg *FileConfig) {
		cfg.ErrorFileName = name
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

// noneLevelColors 返回所有日志级别都不着色的颜色映射。
func noneLevelColors() map[Level]color.ANSI {
	return map[Level]color.ANSI{
		DebugLevel: color.None,
		InfoLevel:  color.None,
		WarnLevel:  color.None,
		ErrorLevel: color.None,
	}
}
