package logx_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/yexuzz/gokit/logx"
	"github.com/yexuzz/gokit/logx/color"
)

// TestDevelopmentUsage 展示开发环境的完整初始化和日志写法。
func TestDevelopmentUsage(t *testing.T) {
	logger, err := logx.Build(
		logx.DevelopmentMode,
		logx.WithFile(logx.FileDir(filepath.Join(t.TempDir(), "logs"))),
		logx.WithLineColor(true),
		logx.WithLevelColor(logx.InfoLevel, color.Cyan),
		logx.WithServiceName("demo-api"),
	)
	if err != nil {
		t.Fatalf("build development logger: %v", err)
	}
	defer logger.Sync()

	old := logx.Default()
	logx.Init(logger)
	defer logx.Init(old)

	ctx := context.Background()
	ctx = logx.WithTraceID(ctx, "trace-dev-001")
	ctx = logx.WithRequestID(ctx, "req-dev-001")

	logx.Debug(ctx, "debug detail", logx.String("module", "order"))
	logx.Info(ctx, "server started",
		logx.LineColor(color.Green),
		logx.String("addr", ":8080"),
		logx.Strings("symbols", []string{"BTCUSDT", "ETHUSDT"}),
	)
	logx.Warnf(ctx, "cache miss: %s", "user:1001")
}

// TestProductionUsage 展示生产环境的完整初始化和日志写法。
func TestProductionUsage(t *testing.T) {
	logger, err := logx.Build(
		logx.ProductionMode,
		logx.WithConsole(true),
		logx.WithColor(false),
		logx.WithFile(
			logx.FileDir(filepath.Join(t.TempDir(), "logs")),
			logx.FileInfo("app"),
			logx.FileError("app-error"),
			logx.FileDateLayout("2006-01-02"),
			logx.FileExt(".log"),
		),
		logx.WithFileRotation(100, 30, 30, true),
		logx.WithCaller(true),
		logx.WithStacktrace(true),
		logx.WithServiceName("coinhub-worker"),
	)
	if err != nil {
		t.Fatalf("build production logger: %v", err)
	}
	defer logger.Sync()

	old := logx.Default()
	logx.Init(logger)
	defer logx.Init(old)

	ctx := context.Background()
	ctx = logx.WithTraceID(ctx, "trace-prod-001")
	ctx = logx.WithRequestID(ctx, "req-prod-001")

	logx.Info(ctx, "order submitted",
		logx.String("order_id", "ord_1001"),
		logx.Int64("uid", 10001),
		logx.Float64("amount", 1288.88),
	)
	logx.Error(ctx, "order failed",
		logx.String("order_id", "ord_1001"),
		logx.Err(context.DeadlineExceeded),
	)
}
