package redis

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/yexuzz/gokit/jwtx"
)

var _ jwtx.Store = (*Store)(nil)

const (
	// defaultPrefix 是 Redis key 的默认前缀.
	defaultPrefix = "jwtx:"
)

// Store 是基于 Redis 的 jwtx.Store 实现.
//
// 它只保存 token 状态, 不保存业务 session 数据.
type Store struct {
	client redis.Cmdable
	prefix string
}

// Option 用于调整 Redis Store.
type Option func(*Store)

// WithPrefix 设置 Redis key 前缀.
func WithPrefix(prefix string) Option {
	return func(store *Store) {
		store.prefix = prefix
	}
}

// NewStore 创建 Redis token 状态存储.
func NewStore(client redis.Cmdable, opts ...Option) *Store {
	store := &Store{
		client: client,
		prefix: defaultPrefix,
	}
	for _, opt := range opts {
		opt(store)
	}
	return store
}

// RevokeSession 写入会话失效标记, 让同一 ssid 下的 access/refresh token 都失效.
func (s *Store) RevokeSession(ctx context.Context, ssid string, ttl time.Duration) error {
	return s.client.Set(ctx, s.revokedKey(ssid), "1", ttl).Err()
}

// IsSessionRevoked 检查会话失效标记是否存在.
func (s *Store) IsSessionRevoked(ctx context.Context, ssid string) (bool, error) {
	n, err := s.client.Exists(ctx, s.revokedKey(ssid)).Result()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

// SaveRefreshTokenID 保存当前 ssid 对应的有效 refresh token jti.
func (s *Store) SaveRefreshTokenID(ctx context.Context, ssid string, tokenID string, ttl time.Duration) error {
	return s.client.Set(ctx, s.refreshKey(ssid), tokenID, ttl).Err()
}

// IsRefreshTokenValid 检查 refresh token jti 是否和服务端保存的一致.
func (s *Store) IsRefreshTokenValid(ctx context.Context, ssid string, tokenID string) (bool, error) {
	val, err := s.client.Get(ctx, s.refreshKey(ssid)).Result()
	if errors.Is(err, redis.Nil) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return val == tokenID, nil
}

// revokedKey 生成会话失效标记的 Redis key.
func (s *Store) revokedKey(ssid string) string {
	return s.prefix + "revoked:" + ssid
}

// refreshKey 生成 refresh token jti 的 Redis key.
func (s *Store) refreshKey(ssid string) string {
	return s.prefix + "refresh:" + ssid
}
