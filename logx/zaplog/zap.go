package zaplog

import (
	"time"

	"github.com/yexuzz/gokit/logx"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// buildZapLogger 根据配置创建 zap 日志器。
func buildZapLogger(cfg Config) (*zap.Logger, error) {
	encoderCfg := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeDuration: zapcore.MillisDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
		EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
		},
	}

	level := zap.NewAtomicLevelAt(zapLevel(cfg.Level))
	cores := make([]zapcore.Core, 0, 2)

	if cfg.Console.Enabled {
		consoleEncoderCfg := encoderCfg
		consoleEncoderCfg.EncodeLevel = levelEncoder(cfg.Console.Color, cfg.Console.LevelColors)
		// 终端输出按运行模式选择 console 或 JSON，便于本地阅读和生产采集。
		cores = append(cores, zapcore.NewCore(
			newEncoder(consoleEncoding(cfg), consoleEncoderCfg),
			consoleWriteSyncer(),
			level,
		))
	}

	if cfg.File.Path != "" {
		fileWriter, err := fileWriteSyncer(cfg.File)
		if err != nil {
			return nil, err
		}
		fileEncoderCfg := encoderCfg
		fileEncoderCfg.EncodeLevel = zapcore.LowercaseLevelEncoder
		// 文件输出固定使用 JSON，避免颜色和终端格式污染日志采集。
		cores = append(cores, zapcore.NewCore(
			newEncoder(JSONEncoding, fileEncoderCfg),
			fileWriter,
			level,
		))
	}

	if len(cores) == 0 {
		encoderCfg.EncodeLevel = zapcore.LowercaseLevelEncoder
		cores = append(cores, zapcore.NewCore(
			newEncoder(JSONEncoding, encoderCfg),
			consoleWriteSyncer(),
			level,
		))
	}

	opts := make([]zap.Option, 0, 4)
	if cfg.AddCaller {
		opts = append(opts, zap.AddCaller(), zap.AddCallerSkip(1))
	}
	if cfg.AddStacktrace {
		opts = append(opts, zap.AddStacktrace(zapcore.ErrorLevel))
	}
	if cfg.Mode == DevelopmentMode {
		opts = append(opts, zap.Development())
	}
	if cfg.ServiceName != "" {
		opts = append(opts, zap.Fields(zap.String("service", cfg.ServiceName)))
	}

	return zap.New(zapcore.NewTee(cores...), opts...), nil
}

// newEncoder 根据编码格式创建 zap 编码器。
func newEncoder(encoding Encoding, cfg zapcore.EncoderConfig) zapcore.Encoder {
	if encoding == ConsoleEncoding {
		return zapcore.NewConsoleEncoder(cfg)
	}
	return zapcore.NewJSONEncoder(cfg)
}

// consoleEncoding 根据运行模式返回终端编码。
func consoleEncoding(cfg Config) Encoding {
	if cfg.Mode == DevelopmentMode {
		return ConsoleEncoding
	}
	return JSONEncoding
}

// zapLevel 将业务日志级别转换为 zap 级别。
func zapLevel(level logx.Level) zapcore.Level {
	switch level {
	case logx.DebugLevel:
		return zapcore.DebugLevel
	case logx.WarnLevel:
		return zapcore.WarnLevel
	case logx.ErrorLevel:
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}
