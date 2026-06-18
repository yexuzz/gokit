package session

import (
	"context"
	"sync"
	"time"
)

var _ Store = (*MemoryStore)(nil)

// MemoryStore 是基于内存 map 的 Store 实现。
//
// 它主要适合单元测试、本地开发和非常轻量的单机场景。
// 生产环境如果有多实例部署，应该使用 Redis 这类共享存储，否则不同实例之间无法共享 Session。
type MemoryStore struct {
	mu   sync.RWMutex
	data map[string]memoryItem
}

type memoryItem struct {
	values    map[string]any
	expiresAt time.Time
}

// NewMemoryStore 创建内存 Session 存储。
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{data: map[string]memoryItem{}}
}

// Init 初始化一个新的内存 Session。
func (s *MemoryStore) Init(ctx context.Context, ssid string, data map[string]any, expiration time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	values := make(map[string]any, len(data))
	for k, v := range data {
		values[k] = v
	}
	s.data[ssid] = memoryItem{
		values:    values,
		expiresAt: time.Now().Add(expiration),
	}
	return nil
}

// Set 写入内存 Session 的单个字段，并顺手延长过期时间。
func (s *MemoryStore) Set(ctx context.Context, ssid string, key string, val any, expiration time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.data[ssid]
	if !ok || time.Now().After(item.expiresAt) {
		delete(s.data, ssid)
		return ErrUnauthorized
	}
	item.values[key] = val
	item.expiresAt = time.Now().Add(expiration)
	s.data[ssid] = item
	return nil
}

// Get 读取内存 Session 的单个字段。
func (s *MemoryStore) Get(ctx context.Context, ssid string, key string) (any, error) {
	s.mu.RLock()
	item, ok := s.data[ssid]
	if !ok {
		s.mu.RUnlock()
		return nil, ErrSessionKeyNotFound
	}
	if time.Now().After(item.expiresAt) {
		s.mu.RUnlock()
		s.mu.Lock()
		delete(s.data, ssid)
		s.mu.Unlock()
		return nil, ErrUnauthorized
	}
	val, ok := item.values[key]
	s.mu.RUnlock()
	if !ok {
		return nil, ErrSessionKeyNotFound
	}
	return val, nil
}

// Del 删除内存 Session 的单个字段。
func (s *MemoryStore) Del(ctx context.Context, ssid string, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.data[ssid]
	if !ok {
		return ErrSessionKeyNotFound
	}
	delete(item.values, key)
	s.data[ssid] = item
	return nil
}

// Destroy 删除整个内存 Session。
func (s *MemoryStore) Destroy(ctx context.Context, ssid string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, ssid)
	return nil
}

var _ Session = (*MemorySession)(nil)

// MemorySession 是一个不依赖 Provider 的轻量 Session 实现。
//
// 它和 MemoryStore 不同：MemorySession 自己持有一份 map，不会签发 JWT，也不会写入 gin.Context。
// 因此它更适合在业务单元测试里直接构造 Session，而不是模拟完整登录流程。
type MemorySession struct {
	mu     sync.RWMutex
	data   map[string]any
	claims Claims
}

// NewMemorySession 创建一个独立的内存 Session。
func NewMemorySession(claims Claims) *MemorySession {
	return &MemorySession{
		data:   map[string]any{},
		claims: claims,
	}
}

// Set 将键值写入当前 MemorySession。
func (m *MemorySession) Set(ctx context.Context, key string, val any) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = val
	return nil
}

// Get 从当前 MemorySession 读取一个值。
func (m *MemorySession) Get(ctx context.Context, key string) (any, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	val, ok := m.data[key]
	if !ok {
		return nil, ErrSessionKeyNotFound
	}
	return val, nil
}

// Del 删除当前 MemorySession 的单个字段。
func (m *MemorySession) Del(ctx context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.data, key)
	return nil
}

// Destroy 清空当前 MemorySession。
func (m *MemorySession) Destroy(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data = map[string]any{}
	return nil
}

// Claims 返回创建 MemorySession 时传入的声明。
func (m *MemorySession) Claims() Claims {
	return m.claims
}
