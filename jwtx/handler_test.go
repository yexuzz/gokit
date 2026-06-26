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

// TestDefaultManager 验证默认 Manager 可以完成登录签发, 校验, 刷新和退出失效.
func TestDefaultManager(t *testing.T) {
	store := newFakeStore()
	ids := []string{"ssid-1", "access-1", "refresh-1", "access-2"}
	manager, err := NewManager[adminPayload](
		WithAccessTokenKey([]byte("access-secret")),
		WithRefreshTokenKey([]byte("refresh-secret")),
		WithStore(store),
		WithIssuer("gokit"),
		WithExpiration(time.Hour, 24*time.Hour),
		WithUserIDExtractor(func(payload adminPayload) string {
			return payload.UID
		}),
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

	pair, err := manager.SetLoginToken(context.Background(), adminPayload{
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
	if !store.userSessions["10001"]["ssid-1"] {
		t.Fatalf("user session not saved: %v", store.userSessions)
	}

	session, err := manager.CheckAccessToken(context.Background(), pair.AccessToken)
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

	newAccessToken, err := manager.RefreshAccessToken(context.Background(), pair.RefreshToken)
	if err != nil {
		t.Fatalf("refresh access token: %v", err)
	}
	if newAccessToken.SSID != "ssid-1" {
		t.Fatalf("unexpected refreshed ssid: %s", newAccessToken.SSID)
	}
	if newAccessToken.TokenID != "access-2" {
		t.Fatalf("unexpected refreshed access token id: %s", newAccessToken.TokenID)
	}
	refreshedSession, err := manager.CheckAccessToken(context.Background(), newAccessToken.Token)
	if err != nil {
		t.Fatalf("check refreshed access token: %v", err)
	}
	if refreshedSession.TokenID != "access-2" {
		t.Fatalf("unexpected refreshed access token id: %s", refreshedSession.TokenID)
	}
	if refreshedSession.Payload.UID != "10001" {
		t.Fatalf("unexpected refreshed payload: %#v", refreshedSession.Payload)
	}

	if err = manager.ClearUserSession(context.Background(), refreshedSession.UserID, newAccessToken.SSID); err != nil {
		t.Fatalf("clear user session: %v", err)
	}
	if _, err = manager.CheckAccessToken(context.Background(), newAccessToken.Token); !errors.Is(err, ErrSessionRevoked) {
		t.Fatalf("expect revoked error, got %v", err)
	}
	if store.userSessions["10001"]["ssid-1"] {
		t.Fatalf("user session not removed: %v", store.userSessions)
	}

	pair, err = manager.SetLoginToken(context.Background(), adminPayload{UID: "10002"})
	if err != nil {
		t.Fatalf("set second login token: %v", err)
	}
	if err = manager.ClearToken(context.Background(), pair.AccessToken); err != nil {
		t.Fatalf("clear token: %v", err)
	}
	if _, err = manager.CheckAccessToken(context.Background(), pair.AccessToken); !errors.Is(err, ErrSessionRevoked) {
		t.Fatalf("expect revoked error, got %v", err)
	}
}

// TestClearUserSessions 验证可以按用户退出全部设备会话.
func TestClearUserSessions(t *testing.T) {
	store := newFakeStore()
	ids := []string{"ssid-web", "access-web", "refresh-web", "ssid-phone", "access-phone", "refresh-phone"}
	manager, err := NewManager[adminPayload](
		WithAccessTokenKey([]byte("access-secret")),
		WithRefreshTokenKey([]byte("refresh-secret")),
		WithStore(store),
		WithUserIDExtractor(func(payload adminPayload) string {
			return payload.UID
		}),
		WithSSIDGenerator(func() string {
			return popID(&ids)
		}),
		WithTokenIDGenerator(func() string {
			return popID(&ids)
		}),
	)
	if err != nil {
		t.Fatalf("new manager: %v", err)
	}

	web, err := manager.SetLoginToken(context.Background(), adminPayload{UID: "10001", Username: "admin"})
	if err != nil {
		t.Fatalf("set web login token: %v", err)
	}
	phone, err := manager.SetLoginToken(context.Background(), adminPayload{UID: "10001", Username: "admin"})
	if err != nil {
		t.Fatalf("set phone login token: %v", err)
	}
	if len(store.userSessions["10001"]) != 2 {
		t.Fatalf("unexpected user sessions: %v", store.userSessions)
	}

	if err = manager.ClearUserSessions(context.Background(), "10001"); err != nil {
		t.Fatalf("clear user sessions: %v", err)
	}
	if _, err = manager.CheckAccessToken(context.Background(), web.AccessToken); !errors.Is(err, ErrSessionRevoked) {
		t.Fatalf("expect web session revoked, got %v", err)
	}
	if _, err = manager.CheckAccessToken(context.Background(), phone.AccessToken); !errors.Is(err, ErrSessionRevoked) {
		t.Fatalf("expect phone session revoked, got %v", err)
	}
	if len(store.userSessions["10001"]) != 0 {
		t.Fatalf("user sessions not cleared: %v", store.userSessions)
	}
}

// TestRefreshTokenInvalid 验证 refresh token jti 和服务端记录不一致时不能刷新.
func TestRefreshTokenInvalid(t *testing.T) {
	store := newFakeStore()
	manager, err := NewManager[adminPayload](
		WithAccessTokenKey([]byte("access-secret")),
		WithRefreshTokenKey([]byte("refresh-secret")),
		WithStore(store),
	)
	if err != nil {
		t.Fatalf("new handler: %v", err)
	}

	pair, err := manager.SetLoginToken(context.Background(), adminPayload{UID: "10001"})
	if err != nil {
		t.Fatalf("set login token: %v", err)
	}
	store.refresh[pair.SSID] = "another-token-id"

	if _, err = manager.RefreshAccessToken(context.Background(), pair.RefreshToken); !errors.Is(err, ErrRefreshTokenInvalid) {
		t.Fatalf("expect invalid refresh token, got %v", err)
	}
}

// fakeStore 是测试用的 token 状态存储.
type fakeStore struct {
	revoked      map[string]bool
	refresh      map[string]string
	userSessions map[string]map[string]bool
}

// newFakeStore 创建测试用的内存 Store.
func newFakeStore() *fakeStore {
	return &fakeStore{
		revoked:      map[string]bool{},
		refresh:      map[string]string{},
		userSessions: map[string]map[string]bool{},
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

// AddUserSession 记录测试用户和 ssid 的绑定关系.
func (s *fakeStore) AddUserSession(ctx context.Context, userID string, ssid string, ttl time.Duration) error {
	if s.userSessions[userID] == nil {
		s.userSessions[userID] = map[string]bool{}
	}
	s.userSessions[userID][ssid] = true
	return nil
}

// RemoveUserSession 移除测试用户和 ssid 的绑定关系.
func (s *fakeStore) RemoveUserSession(ctx context.Context, userID string, ssid string) error {
	delete(s.userSessions[userID], ssid)
	return nil
}

// ListUserSessions 查询测试用户当前记录的全部 ssid.
func (s *fakeStore) ListUserSessions(ctx context.Context, userID string) ([]string, error) {
	ssids := make([]string, 0, len(s.userSessions[userID]))
	for ssid := range s.userSessions[userID] {
		ssids = append(ssids, ssid)
	}
	return ssids, nil
}

// ClearUserSessions 清空测试用户的全部 ssid.
func (s *fakeStore) ClearUserSessions(ctx context.Context, userID string) error {
	delete(s.userSessions, userID)
	return nil
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
