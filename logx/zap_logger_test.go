package logx

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/yexuzz/gokit/logx/color"
	"go.uber.org/zap/zapcore"
)

// TestZapLoggerImplementsLogger 验证 ZapLogger 实现 logx.Logger 接口。
func TestZapLoggerImplementsLogger(t *testing.T) {
	var _ Logger = (*ZapLogger)(nil)
}

// TestZapFormattedLogs 验证 zap 日志器实现格式化日志接口。
func TestZapFormattedLogs(t *testing.T) {
	logger, err := NewDevelopment(WithConsole(false), WithFile(FileDir(t.TempDir())))
	if err != nil {
		t.Fatalf("new development logger: %v", err)
	}
	defer logger.Sync()
	logger.Debugf(context.Background(), "debug %s", "ok")
	logger.Infof(context.Background(), "info %s", "ok")
	logger.Warnf(context.Background(), "warn %s", "ok")
	logger.Errorf(context.Background(), "error %s", "ok")
}

// TestNewDevelopment 验证开发模式 zap 日志器可以正常创建并写入日志。
func TestNewDevelopment(t *testing.T) {
	logger, err := NewDevelopment(
		WithConsole(false),
		WithFile(FileDir(t.TempDir())),
	)
	if err != nil {
		t.Fatalf("new development logger: %v", err)
	}
	defer logger.Sync()
	logger.Info(context.Background(), "started", LineColor(color.Red), String("app", "test"))
}

// TestNewDevelopmentWithoutConsole 验证关闭终端后开发日志器仍可正常写文件。
func TestNewDevelopmentWithoutConsole(t *testing.T) {
	logger, err := NewDevelopment(WithConsole(false), WithFile(FileDir(t.TempDir())))
	if err != nil {
		t.Fatalf("new development logger: %v", err)
	}
	defer logger.Sync()
	logger.Info(context.Background(), "started", LineColor(color.Blue), String("app", "test"), String("1", "2"))
}

// TestBuild 验证可以按运行模式创建对应环境的日志器。
func TestBuild(t *testing.T) {
	cases := []Mode{"dev", "development", "prod", "production", ""}
	for _, mode := range cases {
		logger, err := Build(mode, WithConsole(false), WithFile(FileDir(t.TempDir())))
		if err != nil {
			t.Fatalf("build logger with mode %q: %v", mode, err)
		}
		logger.Info(context.Background(), "started", String("mode", string(mode)))
		if err := logger.Sync(); err != nil {
			t.Fatalf("sync logger with mode %q: %v", mode, err)
		}
	}
}

// TestPackageCaller 验证包级日志方法记录的是业务调用位置，而不是 logx 内部封装位置。
func TestPackageCaller(t *testing.T) {
	dir := t.TempDir()
	logger, err := NewDevelopment(WithConsole(false), WithFile(FileDir(dir)))
	if err != nil {
		t.Fatalf("new development logger: %v", err)
	}
	old := Default()
	Init(logger)
	defer Init(old)

	writeCallerTestLog()
	if err := logger.Sync(); err != nil {
		t.Fatalf("sync logger: %v", err)
	}

	path := filepath.Join(dir, time.Now().Format(DefaultFileDateLayout), DefaultInfoFileName+DefaultFileExt)
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read log file: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) == 0 {
		t.Fatal("log file should not be empty")
	}
	var entry map[string]any
	if err := json.Unmarshal([]byte(lines[len(lines)-1]), &entry); err != nil {
		t.Fatalf("decode log entry: %v", err)
	}
	caller := fmt.Sprint(entry["caller"])
	if !strings.Contains(caller, "zap_logger_test.go") {
		t.Fatalf("caller should point to test file, got %s", caller)
	}
	if strings.Contains(caller, "logger.go") || strings.Contains(caller, "zap_logger.go") {
		t.Fatalf("caller should not point to logx internals, got %s", caller)
	}
}

// writeCallerTestLog 从独立函数写日志，方便断言 caller 指向真实调用位置。
func writeCallerTestLog() {
	Info(context.Background(), "caller-test")
}

// TestNormalizeZapFields 验证 zap 日志器可以转换 logx 字段和宽松字段。
func TestNormalizeZapFields(t *testing.T) {
	fields := normalizeZapFields(String("name", "gokit"), Err(context.Canceled), LineColor(color.Red))
	if len(fields) != 3 {
		t.Fatalf("want 3 fields, got %d", len(fields))
	}
	if fields[0].Key != "name" {
		t.Fatalf("want name field, got %s", fields[0].Key)
	}
	if fields[1].Key != "error" {
		t.Fatalf("want error field, got %s", fields[1].Key)
	}
	if fields[2].Key != lineColorFieldKey {
		t.Fatalf("want line color metadata field, got %s", fields[2].Key)
	}
}

// TestLineColorEncoder 验证整行颜色编码器会给日志行添加颜色。
func TestLineColorEncoder(t *testing.T) {
	encoder := lineColorEncoder{
		Encoder: zapcore.NewConsoleEncoder(zapcore.EncoderConfig{
			MessageKey: "message",
			LevelKey:   "level",
			EncodeLevel: func(level zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
				enc.AppendString(level.String())
			},
		}),
		colors: map[Level]color.ANSI{
			InfoLevel: color.Red,
		},
		useLevelLineColor: true,
	}
	buf, err := encoder.EncodeEntry(zapcore.Entry{Level: zapcore.InfoLevel, Message: "started"}, nil)
	if err != nil {
		t.Fatalf("encode entry: %v", err)
	}
	defer buf.Free()
	if !strings.HasPrefix(buf.String(), string(color.Red)) {
		t.Fatalf("want line color prefix, got %q", buf.String())
	}
}

// TestOnlyLevelColor 验证可以配置只有指定级别显示整行颜色。
func TestOnlyLevelColor(t *testing.T) {
	cfg := DefaultDevelopmentConfig()
	WithLineColor(true)(&cfg)
	WithOnlyLevelColor(InfoLevel, color.Blue)(&cfg)

	encoder := lineColorEncoder{
		Encoder: zapcore.NewConsoleEncoder(zapcore.EncoderConfig{
			MessageKey: "message",
		}),
		colors:            cfg.Console.LevelColors,
		useLevelLineColor: cfg.Console.LineColor,
	}

	infoBuf, err := encoder.EncodeEntry(zapcore.Entry{Level: zapcore.InfoLevel, Message: "info"}, nil)
	if err != nil {
		t.Fatalf("encode info entry: %v", err)
	}
	defer infoBuf.Free()
	if !strings.HasPrefix(infoBuf.String(), string(color.Blue)) {
		t.Fatalf("want info line color prefix, got %q", infoBuf.String())
	}

	warnBuf, err := encoder.EncodeEntry(zapcore.Entry{Level: zapcore.WarnLevel, Message: "warn"}, nil)
	if err != nil {
		t.Fatalf("encode warn entry: %v", err)
	}
	defer warnBuf.Free()
	if strings.HasPrefix(warnBuf.String(), string(color.Blue)) {
		t.Fatalf("warn should not use info color, got %q", warnBuf.String())
	}
}

// TestLineColorField 验证单条日志颜色字段会覆盖级别颜色并被过滤。
func TestLineColorField(t *testing.T) {
	encoder := lineColorEncoder{
		Encoder: zapcore.NewConsoleEncoder(zapcore.EncoderConfig{
			MessageKey: "message",
			LevelKey:   "level",
			EncodeLevel: func(level zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
				enc.AppendString(string(color.Blue) + level.String() + string(color.Reset))
			},
		}),
		colors: map[Level]color.ANSI{
			InfoLevel: color.Green,
		},
	}
	fields := normalizeZapFields(LineColor(color.Red), String("app", "test"))
	buf, err := encoder.EncodeEntry(zapcore.Entry{Level: zapcore.InfoLevel, Message: "started"}, fields)
	if err != nil {
		t.Fatalf("encode entry: %v", err)
	}
	defer buf.Free()
	if !strings.HasPrefix(buf.String(), string(color.Red)) {
		t.Fatalf("want explicit line color prefix, got %q", buf.String())
	}
	if strings.Contains(buf.String(), lineColorFieldKey) {
		t.Fatalf("line color metadata should be filtered, got %q", buf.String())
	}
	if strings.Contains(buf.String(), string(color.Blue)) {
		t.Fatalf("inner level color should be stripped, got %q", buf.String())
	}
}

// TestMetadataFilterEncoder 验证 JSON 或文件编码器会过滤终端颜色元数据。
func TestMetadataFilterEncoder(t *testing.T) {
	encoder := metadataFilterEncoder{
		Encoder: zapcore.NewJSONEncoder(zapcore.EncoderConfig{
			MessageKey: "message",
		}),
	}
	fields := normalizeZapFields(LineColor(color.Red), String("app", "test"))
	buf, err := encoder.EncodeEntry(zapcore.Entry{Message: "started"}, fields)
	if err != nil {
		t.Fatalf("encode entry: %v", err)
	}
	defer buf.Free()
	if strings.Contains(buf.String(), lineColorFieldKey) {
		t.Fatalf("metadata should be filtered, got %q", buf.String())
	}
}
