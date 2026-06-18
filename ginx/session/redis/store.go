package redis

import (
	"context"
	"time"

	"github.com/yexuzz/gokit/ginx/session"
)

var _ session.Store = (*Store)(nil)

// IntCmd 表示 Redis 整数类命令结果。
//
// go-redis 的 *redis.IntCmd 实现了这个接口，因此使用 go-redis 时不需要额外适配命令结果。
type IntCmd interface {
	Err() error
}

// BoolCmd 表示 Redis 布尔类命令结果。
//
// go-redis 的 *redis.BoolCmd 实现了这个接口。
type BoolCmd interface {
	Err() error
}

// StringCmd 表示 Redis 字符串类命令结果。
//
// go-redis 的 *redis.StringCmd 实现了这个接口。
type StringCmd interface {
	Result() (string, error)
}

// Client 定义 RedisStore 需要的最小 Redis 能力。
//
// 这个接口刻意不直接引用 go-redis，目的是让 session/redis 可以适配不同 Redis 客户端。
// 如果你使用 go-redis，redis.Client、redis.ClusterClient、redis.Ring 通常都能直接满足这个接口。
type Client interface {
	HSet(ctx context.Context, key string, values ...any) IntCmd
	HGet(ctx context.Context, key string, field string) StringCmd
	HDel(ctx context.Context, key string, fields ...string) IntCmd
	Del(ctx context.Context, keys ...string) IntCmd
	Expire(ctx context.Context, key string, expiration time.Duration) BoolCmd
}

// Store 是基于 Redis Hash 的 Session 存储实现。
//
// 每个 Session 会被保存成一个 hash：
//
//	key = prefix + ssid
//	field = 业务字段名
//	value = 业务字段值
//
// 注意：Redis 客户端会把 any 转成 Redis 可保存的类型。复杂对象建议由业务提前序列化成 JSON 字符串。
type Store struct {
	client Client
	prefix string
}

// Option 用于调整 Redis Store。
type Option func(*Store)

// WithPrefix 设置 Redis key 前缀。
func WithPrefix(prefix string) Option {
	return func(s *Store) {
		s.prefix = prefix
	}
}

// NewStore 创建 Redis Session Store。
func NewStore(client Client, opts ...Option) *Store {
	s := &Store{
		client: client,
		prefix: "session:",
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// Init 初始化 Redis Hash，并设置过期时间。
func (s *Store) Init(ctx context.Context, ssid string, data map[string]any, expiration time.Duration) error {
	key := s.key(ssid)
	if len(data) > 0 {
		if err := s.client.HSet(ctx, key, data).Err(); err != nil {
			return err
		}
	}
	return s.client.Expire(ctx, key, expiration).Err()
}

// Set 写入 Redis Hash 的单个字段，并延长整个 Session 的过期时间。
func (s *Store) Set(ctx context.Context, ssid string, field string, val any, expiration time.Duration) error {
	key := s.key(ssid)
	if err := s.client.HSet(ctx, key, field, val).Err(); err != nil {
		return err
	}
	return s.client.Expire(ctx, key, expiration).Err()
}

// Get 读取 Redis Hash 的单个字段。
func (s *Store) Get(ctx context.Context, ssid string, field string) (any, error) {
	val, err := s.client.HGet(ctx, s.key(ssid), field).Result()
	if err != nil {
		return nil, err
	}
	return val, nil
}

// Del 删除 Redis Hash 的单个字段。
func (s *Store) Del(ctx context.Context, ssid string, field string) error {
	return s.client.HDel(ctx, s.key(ssid), field).Err()
}

// Destroy 删除整个 Redis Session。
func (s *Store) Destroy(ctx context.Context, ssid string) error {
	return s.client.Del(ctx, s.key(ssid)).Err()
}

func (s *Store) key(ssid string) string {
	if s.prefix == "" {
		return ssid
	}
	return s.prefix + ssid
}
