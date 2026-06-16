package color

import "github.com/yexuzz/gokit/logx"

// ANSI 表示终端 ANSI 颜色控制符。
type ANSI string

const (
	// None 表示不添加颜色。
	None ANSI = ""
	// Blue 表示蓝色终端文本。
	Blue ANSI = "\x1b[34m"
	// Green 表示绿色终端文本。
	Green ANSI = "\x1b[32m"
	// Yellow 表示黄色终端文本。
	Yellow ANSI = "\x1b[33m"
	// Red 表示红色终端文本。
	Red ANSI = "\x1b[31m"
	// Magenta 表示紫色终端文本。
	Magenta ANSI = "\x1b[35m"
	// Cyan 表示青色终端文本。
	Cyan ANSI = "\x1b[36m"
)

// Reset 表示重置终端颜色的 ANSI 控制符。
const Reset ANSI = "\x1b[0m"

// DefaultLevelColors 返回默认日志级别颜色映射。
func DefaultLevelColors() map[logx.Level]ANSI {
	return map[logx.Level]ANSI{
		logx.DebugLevel: Blue,
		logx.InfoLevel:  Green,
		logx.WarnLevel:  Yellow,
		logx.ErrorLevel: Red,
	}
}
