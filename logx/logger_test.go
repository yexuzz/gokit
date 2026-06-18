package logx

import (
	"context"
	"testing"

	"github.com/yexuzz/gokit/logx/color"
)

// captureLogger 记录测试期间收到的日志内容。
type captureLogger struct {
	level  string
	msg    string
	fields []Field
}

// Debug 记录测试调试日志。
func (l *captureLogger) Debug(ctx context.Context, msg string, fields ...Field) {
	l.level = string(DebugLevel)
	l.msg = msg
	l.fields = fields
}

// Info 记录测试普通日志。
func (l *captureLogger) Info(ctx context.Context, msg string, fields ...Field) {
	l.level = string(InfoLevel)
	l.msg = msg
	l.fields = fields
}

// Warn 记录测试警告日志。
func (l *captureLogger) Warn(ctx context.Context, msg string, fields ...Field) {
	l.level = string(WarnLevel)
	l.msg = msg
	l.fields = fields
}

// Error 记录测试错误日志。
func (l *captureLogger) Error(ctx context.Context, msg string, fields ...Field) {
	l.level = string(ErrorLevel)
	l.msg = msg
	l.fields = fields
}

// TestNormalizeFields 验证宽松字段入参可以转换为 logx 字段。
func TestNormalizeFields(t *testing.T) {
	lineColor, fields := normalizeFields(LineColor(color.Red), String("name", "gokit"), Err(context.Canceled))
	if lineColor != color.Red {
		t.Fatalf("want red line color, got %q", lineColor)
	}
	if len(fields) != 2 {
		t.Fatalf("want 2 fields, got %d", len(fields))
	}
	if fields[0].Key != "name" || fields[0].Value != "gokit" {
		t.Fatalf("want name field, got %#v", fields[0])
	}
}

// TestAsLogFunc 验证 logx 日志器可以转换为函数式日志出口。
func TestAsLogFunc(t *testing.T) {
	logger := &captureLogger{}
	fn := AsLogFunc(logger)
	fn(context.Background(), "error", "failed", "err", context.Canceled)
	if logger.level != string(ErrorLevel) {
		t.Fatalf("want error level, got %s", logger.level)
	}
	if len(logger.fields) != 1 {
		t.Fatalf("want 1 loose field, got %d", len(logger.fields))
	}
	field := logger.fields[0]
	if field.Key != "err" || field.Value != context.Canceled {
		t.Fatalf("want err field, got %#v", field)
	}
}

// TestDefaultLogFunc 验证默认日志器可以作为函数式日志出口直接使用。
func TestDefaultLogFunc(t *testing.T) {
	old := Default()
	Init(NewNop())
	defer Init(old)

	fn := DefaultLogFunc()
	fn(context.Background(), "info", "ready", "app", "test")
}

// TestFormattedLogs 验证包级格式化日志方法可以直接调用。
func TestFormattedLogs(t *testing.T) {
	old := Default()
	logger := &captureLogger{}
	Init(logger)
	defer Init(old)

	Infof(context.Background(), "info %s", "ok")
	if logger.level != string(InfoLevel) {
		t.Fatalf("want info level, got %s", logger.level)
	}
	if logger.msg != "info ok" {
		t.Fatalf("want formatted message, got %s", logger.msg)
	}
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
