package jwtx

import (
	"context"
	"errors"
	"testing"
	"time"
)

// adminPayload 模拟业务系统自定义的登录数据.
type adminPayload struct {
	UID      string `json:"uid"`
	Username string `json:"username"`
}

// TestDefaultHandler 验证默认 Handler 可以完成登录签发, 校验, 刷新和退出失效.
func TestDefaultHandler(t *testing.T) {
	store := newFakeStore()
	ids := []string{"ssid-1", "access-1", "refresh-1", "access-2"}
	handler, err := NewHandler[adminPayload](
		WithAccessTokenKey([]byte("access-secret")),
		WithRefreshTokenKey([]byte("refresh-secret")),
		WithStore(store),
		WithIssuer("gokit"),
		WithExpiration(time.Hour, 24*time.Hour),
		WithNow(func() time.Time {
			return time.Date(2099, 6, 23, 12, 0, 0, 0, time.UTC)
		}),
		WithSSIDGenerator(func() string {
			return popID(&ids)
		}),
		WithTokenIDGenerator(func() string {
			return popID(&ids)
		}),
	)
	if err != nil {
		t.Fatalf("new handler: %v", err)
	}

	pair, err := handler.SetLoginToken(context.Background(), adminPayload{
		UID:      "10001",
		Username: "admin",
	}, WithUserAgent("unit-test"))
	if err != nil {
		t.Fatalf("set login token: %v", err)
	}
	if pair.SSID != "ssid-1" {
		t.Fatalf("unexpected ssid: %s", pair.SSID)
	}
	if store.refresh["ssid-1"] != "refresh-1" {
		t.Fatalf("refresh token id not saved: %v", store.refresh)
	}

	session, err := handler.CheckAccessToken(context.Background(), pair.AccessToken)
	if err != nil {
		t.Fatalf("check access token: %v", err)
	}
	if session.Payload.UID != "10001" || session.Payload.Username != "admin" {
		t.Fatalf("unexpected payload: %#v", session.Payload)
	}
	if session.SSID != "ssid-1" || session.TokenID != "access-1" {
		t.Fatalf("unexpected session: %#v", session)
	}
	if session.UserAgent != "unit-test" {
		t.Fatalf("unexpected user agent: %s", session.UserAgent)
	}

	newAccessToken, err := handler.RefreshAccessToken(context.Background(), pair.RefreshToken)
	if err != nil {
		t.Fatalf("refresh access token: %v", err)
	}
	refreshedSession, err := handler.CheckAccessToken(context.Background(), newAccessToken)
	if err != nil {
		t.Fatalf("check refreshed access token: %v", err)
	}
	if refreshedSession.TokenID != "access-2" {
		t.Fatalf("unexpected refreshed access token id: %s", refreshedSession.TokenID)
	}
	if refreshedSession.Payload.UID != "10001" {
		t.Fatalf("unexpected refreshed payload: %#v", refreshedSession.Payload)
	}

	if err = handler.ClearToken(context.Background(), newAccessToken); err != nil {
		t.Fatalf("clear token: %v", err)
	}
	if _, err = handler.CheckAccessToken(context.Background(), newAccessToken); !errors.Is(err, ErrSessionRevoked) {
		t.Fatalf("expect revoked error, got %v", err)
	}
}

// TestRefreshTokenInvalid 验证 refresh token jti 和服务端记录不一致时不能刷新.
func TestRefreshTokenInvalid(t *testing.T) {
	store := newFakeStore()
	handler, err := NewHandler[adminPayload](
		WithAccessTokenKey([]byte("access-secret")),
		WithRefreshTokenKey([]byte("refresh-secret")),
		WithStore(store),
	)
	if err != nil {
		t.Fatalf("new handler: %v", err)
	}

	pair, err := handler.SetLoginToken(context.Background(), adminPayload{UID: "10001"})
	if err != nil {
		t.Fatalf("set login token: %v", err)
	}
	store.refresh[pair.SSID] = "another-token-id"

	if _, err = handler.RefreshAccessToken(context.Background(), pair.RefreshToken); !errors.Is(err, ErrRefreshTokenInvalid) {
		t.Fatalf("expect invalid refresh token, got %v", err)
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

// RevokeSession 记录测试会话已经失效.
func (s *fakeStore) RevokeSession(ctx context.Context, ssid string, ttl time.Duration) error {
	s.revoked[ssid] = true
	return nil
}

// IsSessionRevoked 检查测试会话是否已经失效.
func (s *fakeStore) IsSessionRevoked(ctx context.Context, ssid string) (bool, error) {
	return s.revoked[ssid], nil
}

// SaveRefreshTokenID 保存测试会话当前有效的 refresh token jti.
func (s *fakeStore) SaveRefreshTokenID(ctx context.Context, ssid string, tokenID string, ttl time.Duration) error {
	s.refresh[ssid] = tokenID
	return nil
}

// IsRefreshTokenValid 检查测试 refresh token jti 是否仍有效.
func (s *fakeStore) IsRefreshTokenValid(ctx context.Context, ssid string, tokenID string) (bool, error) {
	return s.refresh[ssid] == tokenID, nil
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
