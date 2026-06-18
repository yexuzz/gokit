package session

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

// TestProviderWithMemoryStore 验证默认 Provider 可以基于内存存储创建并读取 Session。
func TestProviderWithMemoryStore(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx, recorder := newTestContext(http.MethodPost, "/login")
	provider := NewProvider(
		NewMemoryStore(),
		"secret",
		time.Hour,
		WithSSIDGenerator(func() string { return "ssid-1" }),
	)

	sess, err := provider.NewSession(ctx, 123, map[string]string{"role": "admin"}, map[string]any{"device": "pc"})
	if err != nil {
		t.Fatalf("new session: %v", err)
	}
	if recorder.Header().Get("X-Access-Token") == "" {
		t.Fatalf("expect access token in response header")
	}
	if sess.Claims().SSID != "ssid-1" {
		t.Fatalf("unexpected ssid: %s", sess.Claims().SSID)
	}
	role, err := sess.Claims().Get("role")
	if err != nil {
		t.Fatalf("get claim: %v", err)
	}
	if role != "admin" {
		t.Fatalf("unexpected claim role: %v", role)
	}
	device, err := sess.Get(context.Background(), "device")
	if err != nil {
		t.Fatalf("get session value: %v", err)
	}
	if device != "pc" {
		t.Fatalf("unexpected device: %v", device)
	}
}

// TestMiddlewareBuilder 验证登录校验中间件可以解析 header token 并把 Session 写入 gin.Context。
func TestMiddlewareBuilder(t *testing.T) {
	gin.SetMode(gin.TestMode)
	provider := NewProvider(
		NewMemoryStore(),
		"secret",
		time.Hour,
		WithSSIDGenerator(func() string { return "ssid-2" }),
	)
	loginCtx, loginRecorder := newTestContext(http.MethodPost, "/login")
	if _, err := provider.NewSession(loginCtx, 456, nil, nil); err != nil {
		t.Fatalf("new session: %v", err)
	}
	token := loginRecorder.Header().Get("X-Access-Token")

	router := gin.New()
	router.Use((&MiddlewareBuilder{Provider: provider}).Build())
	router.GET("/profile", func(ctx *gin.Context) {
		sess, err := Get(ctx)
		if err != nil {
			t.Fatalf("get session: %v", err)
		}
		ctx.String(http.StatusOK, "%d", sess.Claims().Uid)
	})

	req := httptest.NewRequest(http.MethodGet, "/profile", nil)
	req.Header.Set("Authorization", token)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", recorder.Code)
	}
	if recorder.Body.String() != "456" {
		t.Fatalf("unexpected body: %s", recorder.Body.String())
	}
}

// TestDestroy 验证销毁 Session 后，原 token 不能再通过服务端 Session 校验。
func TestDestroy(t *testing.T) {
	gin.SetMode(gin.TestMode)
	provider := NewProvider(
		NewMemoryStore(),
		"secret",
		time.Hour,
		WithSSIDGenerator(func() string { return "ssid-3" }),
	)
	loginCtx, loginRecorder := newTestContext(http.MethodPost, "/login")
	if _, err := provider.NewSession(loginCtx, 789, nil, nil); err != nil {
		t.Fatalf("new session: %v", err)
	}
	if err := provider.Destroy(loginCtx); err != nil {
		t.Fatalf("destroy session: %v", err)
	}

	readCtx, _ := newTestContext(http.MethodGet, "/profile")
	readCtx.Request.Header.Set("Authorization", loginRecorder.Header().Get("X-Access-Token"))
	if _, err := provider.Get(readCtx); !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("expect unauthorized, got %v", err)
	}
}

// TestDefaultProvider 验证全局默认 Provider 的快捷方法可用。
func TestDefaultProvider(t *testing.T) {
	old := DefaultProvider()
	defer SetDefaultProvider(old)

	provider := NewProvider(NewMemoryStore(), "secret", time.Hour)
	SetDefaultProvider(provider)

	ctx, _ := newTestContext(http.MethodPost, "/login")
	if _, err := NewSession(ctx, 1, nil, nil); err != nil {
		t.Fatalf("new session by default provider: %v", err)
	}
	if _, err := Get(ctx); err != nil {
		t.Fatalf("get session by default provider: %v", err)
	}
}

func newTestContext(method string, path string) (*gin.Context, *httptest.ResponseRecorder) {
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(method, path, nil)
	return ctx, recorder
}
