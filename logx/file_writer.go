package logx

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// consoleWriteSyncer 返回终端日志写入器。
func consoleWriteSyncer() zapcore.WriteSyncer {
	return zapcore.AddSync(os.Stdout)
}

// dailyFileWriter 按日期目录写入日志文件，并在日期变化后自动切换文件。
type dailyFileWriter struct {
	mu      sync.Mutex
	cfg     FileConfig
	name    string
	current string
	writer  *lumberjack.Logger
}

// fileWriteSyncer 创建按日期目录切割并带大小滚动能力的文件日志写入器。
func fileWriteSyncer(cfg FileConfig, name string) (zapcore.WriteSyncer, error) {
	cfg = normalizeFileConfig(cfg)
	writer := &dailyFileWriter{
		cfg:  cfg,
		name: fileNameWithExt(name, cfg.Ext),
	}
	if err := writer.rotate(time.Now()); err != nil {
		return nil, err
	}
	return zapcore.AddSync(writer), nil
}

// Write 写入日志内容，并在日期变化时切换到新的日期目录。
func (w *dailyFileWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if err := w.rotate(time.Now()); err != nil {
		return 0, err
	}
	return w.writer.Write(p)
}

// Sync 刷新当前文件日志缓冲。
func (w *dailyFileWriter) Sync() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.writer == nil {
		return nil
	}
	err := w.writer.Close()
	w.writer = nil
	w.current = ""
	return err
}

// rotate 按日期目录切换底层 lumberjack 写入器。
func (w *dailyFileWriter) rotate(now time.Time) error {
	current := now.Format(w.cfg.DateLayout)
	if w.writer != nil && w.current == current {
		return nil
	}
	if err := cleanupOldDateDirs(w.cfg, now); err != nil {
		return err
	}
	path := filepath.Join(w.cfg.Dir, current, w.name)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	if w.writer != nil {
		_ = w.writer.Close()
	}
	w.current = current
	w.writer = &lumberjack.Logger{
		Filename:   path,
		MaxSize:    w.cfg.MaxSize,
		MaxBackups: w.cfg.MaxBackups,
		MaxAge:     w.cfg.MaxAge,
		Compress:   w.cfg.Compress,
	}
	return nil
}

// cleanupOldDateDirs 清理超过 MaxAge 天的历史日期目录，只删除能按 DateLayout 解析的目录。
func cleanupOldDateDirs(cfg FileConfig, now time.Time) error {
	if cfg.MaxAge <= 0 {
		return nil
	}
	entries, err := os.ReadDir(cfg.Dir)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	cutoff := now.AddDate(0, 0, -cfg.MaxAge)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		dirDate, err := time.ParseInLocation(cfg.DateLayout, entry.Name(), time.Local)
		if err != nil {
			continue
		}
		if dirDate.Before(cutoff) {
			if err := os.RemoveAll(filepath.Join(cfg.Dir, entry.Name())); err != nil {
				return err
			}
		}
	}
	return nil
}

// normalizeFileConfig 补全文件日志配置的默认值。
func normalizeFileConfig(cfg FileConfig) FileConfig {
	if cfg.Dir == "" {
		cfg.Dir = DefaultFileDir
	}
	if cfg.DateLayout == "" {
		cfg.DateLayout = DefaultFileDateLayout
	}
	if cfg.Ext == "" {
		cfg.Ext = DefaultFileExt
	}
	if cfg.InfoFileName == "" {
		cfg.InfoFileName = DefaultInfoFileName
	}
	if cfg.ErrorFileName == "" {
		cfg.ErrorFileName = DefaultErrorFileName
	}
	return cfg
}

// fileNameWithExt 拼接日志文件名和扩展名。
func fileNameWithExt(name string, ext string) string {
	if strings.HasPrefix(ext, ".") {
		return name + ext
	}
	return name + "." + ext
}
