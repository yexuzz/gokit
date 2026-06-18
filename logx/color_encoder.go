package logx

import (
	"strings"

	"github.com/yexuzz/gokit/logx/color"
	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
)

const lineColorFieldKey = "__logx_line_color"

// lineColorEncoder 为整行终端日志增加颜色。
type lineColorEncoder struct {
	zapcore.Encoder
	colors            map[Level]color.ANSI
	useLevelLineColor bool
}

// Clone 克隆整行颜色编码器。
func (e lineColorEncoder) Clone() zapcore.Encoder {
	return lineColorEncoder{
		Encoder:           e.Encoder.Clone(),
		colors:            e.colors,
		useLevelLineColor: e.useLevelLineColor,
	}
}

// EncodeEntry 编码日志并按级别给整行添加颜色。
func (e lineColorEncoder) EncodeEntry(entry zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {
	entryColor, cleanFields := splitLineColor(fields)
	buf, err := e.Encoder.EncodeEntry(entry, cleanFields)
	if err != nil {
		return nil, err
	}
	lineColor := entryColor
	if lineColor == color.None && e.useLevelLineColor {
		logLevel := Level(entry.Level.String())
		if entry.Level >= zapcore.ErrorLevel {
			logLevel = ErrorLevel
		}
		lineColor = e.colors[logLevel]
	}
	if lineColor == color.None {
		return buf, nil
	}
	colored := buffer.NewPool().Get()
	colored.AppendString(string(lineColor))
	colored.AppendString(stripANSI(buf.String()))
	colored.AppendString(string(color.Reset))
	buf.Free()
	return colored, nil
}

// metadataFilterEncoder 过滤只用于终端渲染的日志元数据字段。
type metadataFilterEncoder struct {
	zapcore.Encoder
}

// Clone 克隆元数据过滤编码器。
func (e metadataFilterEncoder) Clone() zapcore.Encoder {
	return metadataFilterEncoder{Encoder: e.Encoder.Clone()}
}

// EncodeEntry 编码日志并移除终端渲染元数据。
func (e metadataFilterEncoder) EncodeEntry(entry zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {
	_, cleanFields := splitLineColor(fields)
	return e.Encoder.EncodeEntry(entry, cleanFields)
}

// splitLineColor 拆分单条日志整行颜色和真实结构化字段。
func splitLineColor(fields []zapcore.Field) (color.ANSI, []zapcore.Field) {
	if len(fields) == 0 {
		return color.None, nil
	}
	lineColor := color.None
	cleanFields := make([]zapcore.Field, 0, len(fields))
	for _, field := range fields {
		if field.Key == lineColorFieldKey && field.Type == zapcore.StringType {
			lineColor = color.ANSI(field.String)
			continue
		}
		cleanFields = append(cleanFields, field)
	}
	return lineColor, cleanFields
}

// stripANSI 移除内部字段颜色，避免单条整行颜色被中途 reset 截断。
func stripANSI(text string) string {
	var b strings.Builder
	b.Grow(len(text))
	inEscape := false
	for i := 0; i < len(text); i++ {
		ch := text[i]
		if inEscape {
			if (ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z') {
				inEscape = false
			}
			continue
		}
		if ch == 0x1b {
			inEscape = true
			continue
		}
		b.WriteByte(ch)
	}
	return b.String()
}

// defaultLevelColors 返回默认日志级别颜色映射。
func defaultLevelColors() map[Level]color.ANSI {
	return map[Level]color.ANSI{
		DebugLevel: color.Blue,
		InfoLevel:  color.Green,
		WarnLevel:  color.Yellow,
		ErrorLevel: color.Red,
	}
}

// levelEncoder 返回日志级别编码器，并按需为终端输出添加颜色。
func levelEncoder(enabled bool, colors map[Level]color.ANSI) zapcore.LevelEncoder {
	if !enabled {
		return zapcore.CapitalLevelEncoder
	}
	return func(level zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
		name := strings.ToUpper(level.String())
		logLevel := Level(level.String())
		if level >= zapcore.ErrorLevel {
			logLevel = ErrorLevel
		}
		levelColor := colors[logLevel]
		if levelColor == color.None {
			enc.AppendString(name)
			return
		}
		enc.AppendString(string(levelColor) + name + string(color.Reset))
	}
}
