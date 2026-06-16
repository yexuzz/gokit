package zaplog

import (
	"context"
	"testing"

	"github.com/yexuzz/gokit/logx"
)

// TestLoggerImplementsLogxLogger 验证 zaplog.Logger 实现 logx.Logger 接口。
func TestLoggerImplementsLogxLogger(t *testing.T) {
	var _ logx.Logger = (*Logger)(nil)
}

// TestNewDevelopment 验证开发模式 zap 日志器可以正常创建并写入日志。
func TestNewDevelopment(t *testing.T) {
	logger, err := NewDevelopment(WithConsole(false))
	if err != nil {
		t.Fatalf("new development logger: %v", err)
	}
	logger.Info(context.Background(), "started", logx.String("app", "test"))
}

// TestNormalizeFields 验证 zaplog 可以转换 logx 字段和宽松字段。
func TestNormalizeFields(t *testing.T) {
	fields := normalizeFields(logx.String("name", "gokit"), "age", 18, context.Canceled)
	if len(fields) != 3 {
		t.Fatalf("want 3 fields, got %d", len(fields))
	}
}
