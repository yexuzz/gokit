package logx

import (
	"context"
	"testing"
)

// TestNormalizeFields 验证宽松字段入参可以转换为 logx 字段。
func TestNormalizeFields(t *testing.T) {
	fields := normalizeFields(String("name", "gokit"), "age", 18, context.Canceled)
	if len(fields) != 3 {
		t.Fatalf("want 3 fields, got %d", len(fields))
	}
}

// TestAsLogFunc 验证 logx 日志器可以转换为函数式日志出口。
func TestAsLogFunc(t *testing.T) {
	logger := NewNop()
	fn := AsLogFunc(logger)
	fn(context.Background(), "error", "failed", "err", context.Canceled)
}

// TestDefaultLogFunc 验证默认日志器可以作为函数式日志出口直接使用。
func TestDefaultLogFunc(t *testing.T) {
	old := Default()
	Init(NewNop())
	defer Init(old)

	fn := DefaultLogFunc()
	fn(context.Background(), "info", "ready", "app", "test")
}

// TestContextFields 验证日志器会从上下文提取 trace_id 和 request_id。
func TestContextFields(t *testing.T) {
	ctx := context.Background()
	ctx = WithTraceID(ctx, "trace-1")
	ctx = WithRequestID(ctx, "req-1")

	fields := ContextFields(ctx)
	if len(fields) != 2 {
		t.Fatalf("want 2 fields, got %d", len(fields))
	}
}
