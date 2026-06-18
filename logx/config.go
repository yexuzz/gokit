package logx

import (
	"github.com/yexuzz/gokit/logx/color"
)

// Mode 表示 zap 日志器的运行模式。
type Mode string

const (
	// DevelopmentMode 表示开发模式，默认输出更适合终端阅读的彩色文本。
	DevelopmentMode Mode = "dev"
	// ProductionMode 表示生产模式，默认输出更适合采集系统处理的 JSON。
	ProductionMode Mode = "prod"
)

// Encoding 表示 zap 日志编码格式。
type Encoding string

const (
	// ConsoleEncoding 表示面向终端阅读的文本编码。
	ConsoleEncoding Encoding = "console"
	// JSONEncoding 表示面向机器采集的 JSON 编码。
	JSONEncoding Encoding = "json"
)

const (
	// DefaultFileDir 表示启用文件输出但未指定目录时使用的默认日志目录。
	DefaultFileDir = "./logs"
	// DefaultFileDateLayout 表示日期目录的默认格式。
	DefaultFileDateLayout = "2006-01-02"
	// DefaultFileExt 表示日志文件的默认扩展名。
	DefaultFileExt = ".log"

	// DefaultInfoFileName 表示普通日志的默认文件名。
	DefaultInfoFileName = "app"
	// DefaultErrorFileName 表示错误日志的默认文件名。
	DefaultErrorFileName = "app-error"
)

// ConsoleConfig 表示终端输出配置。
type ConsoleConfig struct {
	Enabled     bool                 // 是否开启终端输出
	Color       bool                 // 是否开启终端级别颜色
	LineColor   bool                 // 是否按日志级别给整行终端日志着色
	LevelColors map[Level]color.ANSI // 各日志级别对应的终端颜色
}

// FileConfig 表示文件输出配置。
type FileConfig struct {
	Enabled       bool   // 是否开启文件输出
	Dir           string // 日志根目录
	DateLayout    string // 日期目录格式
	Ext           string // 日志文件扩展名
	InfoFileName  string // 普通日志文件名
	ErrorFileName string // 错误日志文件名
	MaxSize       int    // 单个日志文件最大 MB
	MaxBackups    int    // 保留旧日志文件数量
	MaxAge        int    // 保留旧日志文件天数
	Compress      bool   // 是否压缩旧日志文件
}

// Config 表示 zap 日志器完整配置。
type Config struct {
	Mode          Mode          // 运行模式
	Level         Level         // 最小输出级别
	Console       ConsoleConfig // 终端输出配置
	File          FileConfig    // 文件输出配置
	AddCaller     bool          // 是否记录调用位置
	AddStacktrace bool          // 是否记录错误堆栈
	ServiceName   string        // 服务名称，会作为公共字段输出
}

// DefaultConfig 按运行模式返回日志器默认配置。
func DefaultConfig(mode Mode) Config {
	mode = normalizeMode(mode)
	cfg := Config{
		Mode:  mode,
		Level: InfoLevel,
		Console: ConsoleConfig{
			Enabled:     true,
			Color:       false,
			LevelColors: defaultLevelColors(),
		},
		File:          defaultFileConfig(),
		AddCaller:     true,
		AddStacktrace: true,
	}
	if mode == DevelopmentMode {
		cfg.Level = DebugLevel
		cfg.Console.Color = true
	}
	return cfg
}

// DefaultDevelopmentConfig 返回适合本地开发和终端观察的默认配置。
func DefaultDevelopmentConfig() Config {
	return DefaultConfig(DevelopmentMode)
}

// DefaultProductionConfig 返回适合生产环境和日志采集的默认配置。
func DefaultProductionConfig() Config {
	return DefaultConfig(ProductionMode)
}

// defaultFileConfig 返回文件日志的默认配置。
func defaultFileConfig() FileConfig {
	return FileConfig{
		Enabled:       true,
		Dir:           DefaultFileDir,
		DateLayout:    DefaultFileDateLayout,
		Ext:           DefaultFileExt,
		InfoFileName:  DefaultInfoFileName,
		ErrorFileName: DefaultErrorFileName,
		MaxSize:       100,
		MaxBackups:    30,
		MaxAge:        30,
		Compress:      true,
	}
}
