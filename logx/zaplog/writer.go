package zaplog

import (
	"os"
	"path/filepath"

	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// consoleWriteSyncer 返回终端日志写入器。
func consoleWriteSyncer() zapcore.WriteSyncer {
	return zapcore.AddSync(os.Stdout)
}

// fileWriteSyncer 创建日志目录并返回带滚动能力的文件日志写入器。
func fileWriteSyncer(cfg FileConfig) (zapcore.WriteSyncer, error) {
	if dir := filepath.Dir(cfg.Path); dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, err
		}
	}
	return zapcore.AddSync(&lumberjack.Logger{
		Filename:   cfg.Path,
		MaxSize:    cfg.MaxSize,
		MaxBackups: cfg.MaxBackups,
		MaxAge:     cfg.MaxAge,
		Compress:   cfg.Compress,
	}), nil
}
