package jwtx

import (
	"context"
	"time"
)

var _ Store = NoopStore{}

// NoopStore 是不保存任何服务端状态的 Store 实现.
//
// 它适合纯 JWT 场景. 使用它时 ClearToken 不会让已签发 token 主动失效,
// refresh token 也只依赖 JWT 自身签名和过期时间.
type NoopStore struct{}

// RevokeSession 忽略会话失效写入.
func (NoopStore) RevokeSession(ctx context.Context, ssid string, ttl time.Duration) error {
	return nil
}

// IsSessionRevoked 固定返回 false.
func (NoopStore) IsSessionRevoked(ctx context.Context, ssid string) (bool, error) {
	return false, nil
}

// SaveRefreshTokenID 忽略 refresh token jti 写入.
func (NoopStore) SaveRefreshTokenID(ctx context.Context, ssid string, tokenID string, ttl time.Duration) error {
	return nil
}

// IsRefreshTokenValid 固定返回 true.
func (NoopStore) IsRefreshTokenValid(ctx context.Context, ssid string, tokenID string) (bool, error) {
	return true, nil
}
