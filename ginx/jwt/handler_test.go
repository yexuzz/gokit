package jwt

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yexuzz/gokit/jwtx"
)

// adminPayload 模拟业务系统放入 JWT 的后台用户信息.
type adminPayload struct {
	UID      string `json:"uid"`
	Username string `json:"username"`
}

// TestHandler 验证 Gin JWT 适配层可以完成登录写 token, 中间件写 session, 刷新 token 和退出失效.
func TestHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	store := newFakeStore()
	ids := []string{"ssid-1", "access-1", "refresh-1", "access-2"}
	manager := jwtx.MustNewManager[adminPayload](
		jwtx.WithAccessTokenKey([]byte("access-secret")),
		jwtx.WithRefreshTokenKey([]byte("refresh-secret")),
		jwtx.WithStore(store),
		jwtx.WithNow(func() time.Time {
			return time.Date(2099, 6, 25, 12, 0, 0, 0, time.UTC)
		}),
		jwtx.WithSSIDGenerator(func() string {
			return popID(&ids)
		}),
		jwtx.WithTokenIDGenerator(func() string {
			return popID(&ids)
		}),
	)
	handler := NewHandler(manager)

	loginCtx, _ := newTestContext(http.MethodPost, "/login", "", "unit-test")
	ssid, err := handler.SetLoginToken(loginCtx, adminPayload{
		UID:      "10001",
		Username: "admin",
	})
	if err != nil {
		t.Fatalf("set login token: %v", err)
	}
	if ssid != "ssid-1" {
		t.Fatalf("unexpected ssid: %s", ssid)
	}
	accessToken := loginCtx.Writer.Header().Get(DefaultAccessTokenHeader)
	refreshToken := loginCtx.Writer.Header().Get(DefaultRefreshTokenHeader)
	if accessToken == "" || refreshToken == "" {
		t.Fatalf("token header not written, access=%q refresh=%q", accessToken, refreshToken)
	}

	authCtx, _ := newTestContext(http.MethodGet, "/profile", "Bearer "+accessToken, "unit-test")
	NewLoginMiddlewareBuilder(handler).Build()(authCtx)
	if authCtx.IsAborted() {
		t.Fatal("middleware should allow valid token")
	}
	session, ok := Session[adminPayload](authCtx)
	if !ok {
		t.Fatal("session not saved into gin context")
	}
	if session.SSID != "ssid-1" || session.Payload.UID != "10001" {
		t.Fatalf("unexpected session: %#v", session)
	}

	refreshCtx, _ := newTestContext(http.MethodPost, "/refresh", refreshToken, "unit-test")
	newAccessToken, err := handler.RefreshToken(refreshCtx)
	if err != nil {
		t.Fatalf("refresh token: %v", err)
	}
	if newAccessToken.SSID != "ssid-1" || newAccessToken.TokenID != "access-2" {
		t.Fatalf("unexpected refreshed token: %#v", newAccessToken)
	}
	if refreshCtx.Writer.Header().Get(DefaultAccessTokenHeader) != newAccessToken.Token {
		t.Fatal("new access token header not written")
	}

	if err = handler.ClearToken(authCtx); err != nil {
		t.Fatalf("clear token: %v", err)
	}
	if _, err = manager.CheckAccessToken(context.Background(), newAccessToken.Token); !errors.Is(err, jwtx.ErrSessionRevoked) {
		t.Fatalf("expect revoked session, got %v", err)
	}
}

// TestMiddlewareUserAgentMismatch 验证默认会校验 token 中记录的 User-Agent.
func TestMiddlewareUserAgentMismatch(t *testing.T) {
	gin.SetMode(gin.TestMode)

	manager := jwtx.MustNewManager[adminPayload](
		jwtx.WithAccessTokenKey([]byte("access-secret")),
		jwtx.WithRefreshTokenKey([]byte("refresh-secret")),
	)
	handler := NewHandler(manager)

	loginCtx, _ := newTestContext(http.MethodPost, "/login", "", "old-agent")
	if _, err := handler.SetLoginToken(loginCtx, adminPayload{UID: "10001"}); err != nil {
		t.Fatalf("set login token: %v", err)
	}

	authCtx, recorder := newTestContext(
		http.MethodGet,
		"/profile",
		"Bearer "+loginCtx.Writer.Header().Get(DefaultAccessTokenHeader),
		"new-agent",
	)
	NewLoginMiddlewareBuilder(handler).Build()(authCtx)
	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expect unauthorized, got %d", recorder.Code)
	}
}

// TestLoginMiddlewareBuilderIgnorePaths 验证登录态中间件 Builder 可以配置不需要校验的路径.
func TestLoginMiddlewareBuilderIgnorePaths(t *testing.T) {
	gin.SetMode(gin.TestMode)

	manager := jwtx.MustNewManager[adminPayload](
		jwtx.WithAccessTokenKey([]byte("access-secret")),
		jwtx.WithRefreshTokenKey([]byte("refresh-secret")),
	)
	handler := NewHandler(manager)
	middleware := NewLoginMiddlewareBuilder(handler).
		IgnorePaths("/login", "/health").
		Build()

	ctx, recorder := newTestContext(http.MethodPost, "/login", "", "")
	middleware(ctx)
	if ctx.IsAborted() {
		t.Fatal("ignored path should not be aborted")
	}
	if recorder.Code == http.StatusUnauthorized {
		t.Fatal("ignored path should not return unauthorized")
	}
}

// fakeStore 是测试用的 token 状态存储.
type fakeStore struct {
	revoked map[string]bool
	refresh map[string]string
}

// newFakeStore 创建测试用的内存 Store.
func newFakeStore() *fakeStore {
	return &fakeStore{
		revoked: map[string]bool{},
		refresh: map[string]string{},
	}
}

// RevokeSession 记录测试会话已经主动失效.
func (s *fakeStore) RevokeSession(ctx context.Context, ssid string, ttl time.Duration) error {
	s.revoked[ssid] = true
	return nil
}

// IsSessionRevoked 检查测试会话是否已经主动失效.
func (s *fakeStore) IsSessionRevoked(ctx context.Context, ssid string) (bool, error) {
	return s.revoked[ssid], nil
}

// SaveRefreshTokenID 保存测试会话当前有效的 refresh token jti.
func (s *fakeStore) SaveRefreshTokenID(ctx context.Context, ssid string, tokenID string, ttl time.Duration) error {
	s.refresh[ssid] = tokenID
	return nil
}

// IsRefreshTokenValid 检查测试 refresh token jti 是否仍然有效.
func (s *fakeStore) IsRefreshTokenValid(ctx context.Context, ssid string, tokenID string) (bool, error) {
	return s.refresh[ssid] == tokenID, nil
}

// newTestContext 创建带 Authorization 和 User-Agent 的 Gin 测试上下文.
func newTestContext(method string, path string, token string, userAgent string) (*gin.Context, *httptest.ResponseRecorder) {
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	req := httptest.NewRequest(method, path, nil)
	if token != "" {
		req.Header.Set("Authorization", token)
	}
	if userAgent != "" {
		req.Header.Set("User-Agent", userAgent)
	}
	ctx.Request = req
	return ctx, recorder
}

// popID 按顺序弹出测试用固定 ID.
func popID(ids *[]string) string {
	if len(*ids) == 0 {
		return "fallback"
	}
	id := (*ids)[0]
	*ids = (*ids)[1:]
	return id
}
