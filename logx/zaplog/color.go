package zaplog

import (
	"strings"

	"github.com/yexuzz/gokit/logx"
	"github.com/yexuzz/gokit/logx/color"
	"go.uber.org/zap/zapcore"
)

// levelEncoder 返回日志级别编码器，并按需为终端输出添加颜色。
func levelEncoder(enabled bool, colors map[logx.Level]color.ANSI) zapcore.LevelEncoder {
	if !enabled {
		return zapcore.CapitalLevelEncoder
	}
	return func(level zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
		name := strings.ToUpper(level.String())
		logLevel := logx.Level(level.String())
		if level >= zapcore.ErrorLevel {
			logLevel = logx.ErrorLevel
		}
		levelColor := colors[logLevel]
		if levelColor == color.None {
			enc.AppendString(name)
			return
		}
		enc.AppendString(string(levelColor) + name + string(color.Reset))
	}
}
