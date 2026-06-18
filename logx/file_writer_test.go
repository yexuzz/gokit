package logx

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestDailyFileWriterPath 验证文件日志会写入日期目录下的指定日志文件。
func TestDailyFileWriterPath(t *testing.T) {
	cfg := normalizeFileConfig(FileConfig{
		Dir:        t.TempDir(),
		DateLayout: DefaultFileDateLayout,
		Ext:        DefaultFileExt,
		MaxSize:    1,
	})
	writer := &dailyFileWriter{
		cfg:  cfg,
		name: fileNameWithExt(DefaultInfoFileName, cfg.Ext),
	}
	if err := writer.rotate(time.Date(2026, 6, 18, 1, 2, 3, 0, time.Local)); err != nil {
		t.Fatalf("rotate writer: %v", err)
	}
	defer writer.Sync()

	want := filepath.Join(cfg.Dir, "2026-06-18", "app.log")
	if writer.writer == nil || writer.writer.Filename != want {
		t.Fatalf("want writer path %s, got %#v", want, writer.writer)
	}
	if _, err := os.Stat(filepath.Dir(want)); err != nil {
		t.Fatalf("want date dir created: %v", err)
	}
}

// TestFileNameWithExt 验证日志文件扩展名可以带点或不带点。
func TestFileNameWithExt(t *testing.T) {
	if got := fileNameWithExt("info", ".log"); got != "info.log" {
		t.Fatalf("want info.log, got %s", got)
	}
	if got := fileNameWithExt("info", "json"); got != "info.json" {
		t.Fatalf("want info.json, got %s", got)
	}
}

// TestCleanupOldDateDirs 验证旧日期目录会按 MaxAge 清理，非日期目录会保留。
func TestCleanupOldDateDirs(t *testing.T) {
	root := t.TempDir()
	for _, dir := range []string{"2026-06-01", "2026-06-17", "misc"} {
		if err := os.MkdirAll(filepath.Join(root, dir), 0755); err != nil {
			t.Fatalf("create dir %s: %v", dir, err)
		}
	}
	cfg := normalizeFileConfig(FileConfig{
		Dir:        root,
		DateLayout: DefaultFileDateLayout,
		MaxAge:     7,
	})
	now := time.Date(2026, 6, 18, 12, 0, 0, 0, time.Local)
	if err := cleanupOldDateDirs(cfg, now); err != nil {
		t.Fatalf("cleanup old dirs: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "2026-06-01")); !os.IsNotExist(err) {
		t.Fatalf("old date dir should be removed, stat err: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "2026-06-17")); err != nil {
		t.Fatalf("recent date dir should be kept: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "misc")); err != nil {
		t.Fatalf("non-date dir should be kept: %v", err)
	}
}
