package logx

import "testing"

// TestWithFileDefaultConfig 验证启用文件输出但不传配置时使用默认目录和文件名。
func TestWithFileDefaultConfig(t *testing.T) {
	cfg := DefaultProductionConfig()
	WithFile()(&cfg)
	if !cfg.File.Enabled {
		t.Fatal("file output should be enabled")
	}
	if cfg.File.Dir != DefaultFileDir {
		t.Fatalf("want default file dir %s, got %s", DefaultFileDir, cfg.File.Dir)
	}
	if cfg.File.DateLayout != DefaultFileDateLayout {
		t.Fatalf("want default date layout %s, got %s", DefaultFileDateLayout, cfg.File.DateLayout)
	}
	if cfg.File.InfoFileName != DefaultInfoFileName {
		t.Fatalf("want default info name %s, got %s", DefaultInfoFileName, cfg.File.InfoFileName)
	}
	if cfg.File.ErrorFileName != DefaultErrorFileName {
		t.Fatalf("want default error name %s, got %s", DefaultErrorFileName, cfg.File.ErrorFileName)
	}
}

// TestWithoutFile 验证可以显式关闭默认文件日志输出。
func TestWithoutFile(t *testing.T) {
	cfg := DefaultDevelopmentConfig()
	WithoutFile()(&cfg)
	if cfg.File.Enabled {
		t.Fatal("file output should be disabled")
	}
}

// TestDevelopmentConfigEnablesFile 验证开发模式默认同时写入文件日志。
func TestDevelopmentConfigEnablesFile(t *testing.T) {
	cfg := DefaultDevelopmentConfig()
	if !cfg.File.Enabled {
		t.Fatal("development file output should be enabled")
	}
	if cfg.File.Dir != DefaultFileDir {
		t.Fatalf("want default file dir %s, got %s", DefaultFileDir, cfg.File.Dir)
	}
}

// TestDefaultConfigByMode 验证默认配置会按运行模式切换级别和终端颜色。
func TestDefaultConfigByMode(t *testing.T) {
	dev := DefaultConfig("development")
	if dev.Mode != DevelopmentMode || dev.Level != DebugLevel || !dev.Console.Color {
		t.Fatalf("unexpected dev config: %#v", dev)
	}
	prod := DefaultConfig("unknown")
	if prod.Mode != ProductionMode || prod.Level != InfoLevel || prod.Console.Color {
		t.Fatalf("unexpected prod config: %#v", prod)
	}
}

// TestWithFileCustomConfig 验证文件输出可以覆盖目录和文件名。
func TestWithFileCustomConfig(t *testing.T) {
	cfg := DefaultProductionConfig()
	WithFile(
		FileDir("./runtime/logs"),
		FileDateLayout("20060102"),
		FileExt("json"),
		FileInfo("biz"),
		FileError("err"),
	)(&cfg)

	if cfg.File.Dir != "./runtime/logs" {
		t.Fatalf("want custom file dir, got %s", cfg.File.Dir)
	}
	if cfg.File.DateLayout != "20060102" {
		t.Fatalf("want custom date layout, got %s", cfg.File.DateLayout)
	}
	if cfg.File.Ext != "json" {
		t.Fatalf("want custom ext, got %s", cfg.File.Ext)
	}
	if cfg.File.InfoFileName != "biz" {
		t.Fatalf("want custom info name, got %s", cfg.File.InfoFileName)
	}
	if cfg.File.ErrorFileName != "err" {
		t.Fatalf("want custom error name, got %s", cfg.File.ErrorFileName)
	}
}
