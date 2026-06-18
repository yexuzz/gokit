package logx

import (
	"time"

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
		cores = append(cores, newConsoleCore(cfg, encoderCfg, level))
	}

	if cfg.File.Enabled {
		fileCores, err := newFileCores(cfg, encoderCfg)
		if err != nil {
			return nil, err
		}
		cores = append(cores, fileCores...)
	}

	opts := make([]zap.Option, 0, 4)
	if cfg.AddCaller {
		opts = append(opts, zap.AddCaller(), zap.AddCallerSkip(2))
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

	if len(cores) == 0 {
		// 没有开启任何输出时保持静默，避免 WithConsole(false) 仍然写 stdout。
		return zap.NewNop(), nil
	}
	return zap.New(zapcore.NewTee(cores...), opts...), nil
}

// newConsoleCore 创建终端日志输出 core。
func newConsoleCore(cfg Config, encoderCfg zapcore.EncoderConfig, level zap.AtomicLevel) zapcore.Core {
	consoleEncoderCfg := encoderCfg
	// 整行着色时不能再给 level 单独着色，否则 level 后面的 reset 会截断整行颜色。
	consoleEncoderCfg.EncodeLevel = levelEncoder(cfg.Console.Color && !cfg.Console.LineColor, cfg.Console.LevelColors)
	consoleEncoder := newEncoder(consoleEncoding(cfg), consoleEncoderCfg)
	if consoleEncoding(cfg) == ConsoleEncoding {
		consoleEncoder = lineColorEncoder{
			Encoder:           consoleEncoder,
			colors:            cfg.Console.LevelColors,
			useLevelLineColor: cfg.Console.LineColor,
		}
	}
	return zapcore.NewCore(consoleEncoder, consoleWriteSyncer(), level)
}

// newFileCores 创建普通日志和错误日志两个文件输出 core。
func newFileCores(cfg Config, encoderCfg zapcore.EncoderConfig) ([]zapcore.Core, error) {
	fileCfg := normalizeFileConfig(cfg.File)
	infoWriter, err := fileWriteSyncer(fileCfg, fileCfg.InfoFileName)
	if err != nil {
		return nil, err
	}
	errorWriter, err := fileWriteSyncer(fileCfg, fileCfg.ErrorFileName)
	if err != nil {
		return nil, err
	}
	fileEncoderCfg := encoderCfg
	fileEncoderCfg.EncodeLevel = zapcore.LowercaseLevelEncoder
	infoLevel := zap.LevelEnablerFunc(func(level zapcore.Level) bool {
		return level >= zapLevel(cfg.Level) && level < zapcore.ErrorLevel
	})
	errorLevel := zap.LevelEnablerFunc(func(level zapcore.Level) bool {
		return level >= zapcore.ErrorLevel
	})
	return []zapcore.Core{
		zapcore.NewCore(
			metadataFilterEncoder{Encoder: newEncoder(JSONEncoding, fileEncoderCfg)},
			infoWriter,
			infoLevel,
		),
		zapcore.NewCore(
			metadataFilterEncoder{Encoder: newEncoder(JSONEncoding, fileEncoderCfg)},
			errorWriter,
			errorLevel,
		),
	}, nil
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
func zapLevel(level Level) zapcore.Level {
	switch level {
	case DebugLevel:
		return zapcore.DebugLevel
	case WarnLevel:
		return zapcore.WarnLevel
	case ErrorLevel:
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}
