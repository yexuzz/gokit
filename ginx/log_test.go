package ginx

import (
	"context"
	"testing"
)

// captureLog 记录测试期间 ginx 发出的日志事件。
type captureLog struct {
	level  string
	msg    string
	fields []any
}

// TestSetLogFunc 验证 ginx 可以通过函数注入外部日志处理逻辑。
func TestSetLogFunc(t *testing.T) {
	l := &captureLog{}
	SetLogFunc(func(ctx context.Context, level string, msg string, fields ...any) {
		l.level = level
		l.msg = msg
		l.fields = fields
	})
	defer SetLogFunc(nil)

	logError(context.Background(), "failed", "err", "mock")

	if l.level != "error" {
		t.Fatalf("want level error, got %s", l.level)
	}
	if l.msg != "failed" {
		t.Fatalf("want msg failed, got %s", l.msg)
	}
	if len(l.fields) != 2 {
		t.Fatalf("want 2 fields, got %d", len(l.fields))
	}
}
