package zaplog

import (
	"github.com/yexuzz/gokit/logx"
	"github.com/yexuzz/gokit/logx/color"
)

// Mode 表示 zap 日志器的运行模式。
type Mode string

const (
	// DevelopmentMode 表示开发模式，默认输出更适合终端阅读的彩色文本。
	DevelopmentMode Mode = "development"
	// ProductionMode 表示生产模式，默认输出更适合采集系统处理的 JSON。
	ProductionMode Mode = "production"
)

// Encoding 表示 zap 日志编码格式。
type Encoding string

const (
	// ConsoleEncoding 表示面向终端阅读的文本编码。
	ConsoleEncoding Encoding = "console"
	// JSONEncoding 表示面向机器采集的 JSON 编码。
	JSONEncoding Encoding = "json"
)

// ConsoleConfig 表示终端输出配置。
type ConsoleConfig struct {
	Enabled     bool                      // 是否开启终端输出
	Color       bool                      // 是否开启终端级别颜色
	LevelColors map[logx.Level]color.ANSI // 各日志级别对应的终端颜色
}

// FileConfig 表示文件输出配置。
type FileConfig struct {
	Path       string // 日志文件路径
	MaxSize    int    // 单个日志文件最大 MB
	MaxBackups int    // 保留旧日志文件数量
	MaxAge     int    // 保留旧日志文件天数
	Compress   bool   // 是否压缩旧日志文件
}

// Config 表示 zap 日志器完整配置。
type Config struct {
	Mode          Mode          // 运行模式
	Level         logx.Level    // 最小输出级别
	Console       ConsoleConfig // 终端输出配置
	File          FileConfig    // 文件输出配置
	AddCaller     bool          // 是否记录调用位置
	AddStacktrace bool          // 是否记录错误堆栈
	ServiceName   string        // 服务名称，会作为公共字段输出
}

// DefaultDevelopmentConfig 返回适合本地开发和终端观察的默认配置。
func DefaultDevelopmentConfig() Config {
	return Config{
		Mode:  DevelopmentMode,
		Level: logx.DebugLevel,
		Console: ConsoleConfig{
			Enabled:     true,
			Color:       true,
			LevelColors: color.DefaultLevelColors(),
		},
		AddCaller:     true,
		AddStacktrace: true,
	}
}

// DefaultProductionConfig 返回适合生产环境和日志采集的默认配置。
func DefaultProductionConfig() Config {
	return Config{
		Mode:  ProductionMode,
		Level: logx.InfoLevel,
		Console: ConsoleConfig{
			Enabled:     true,
			Color:       false,
			LevelColors: color.DefaultLevelColors(),
		},
		File: FileConfig{
			MaxSize:    100,
			MaxBackups: 30,
			MaxAge:     30,
			Compress:   true,
		},
		AddCaller:     true,
		AddStacktrace: true,
	}
}
