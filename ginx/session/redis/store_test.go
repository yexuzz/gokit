package redis

import (
	"context"
	"testing"
	"time"
)

// TestStore 验证 Redis Store 会按 hash 方式读写 Session 字段。
func TestStore(t *testing.T) {
	client := newFakeClient()
	store := NewStore(client, WithPrefix("test:"))
	ctx := context.Background()

	if err := store.Init(ctx, "ssid", map[string]any{"uid": int64(123)}, time.Hour); err != nil {
		t.Fatalf("init store: %v", err)
	}
	if err := store.Set(ctx, "ssid", "device", "pc", time.Hour); err != nil {
		t.Fatalf("set field: %v", err)
	}
	val, err := store.Get(ctx, "ssid", "device")
	if err != nil {
		t.Fatalf("get field: %v", err)
	}
	if val != "pc" {
		t.Fatalf("unexpected val: %v", val)
	}
	if err = store.Del(ctx, "ssid", "device"); err != nil {
		t.Fatalf("delete field: %v", err)
	}
	if err = store.Destroy(ctx, "ssid"); err != nil {
		t.Fatalf("destroy session: %v", err)
	}
}

type fakeClient struct {
	data map[string]map[string]string
}

func newFakeClient() *fakeClient {
	return &fakeClient{data: map[string]map[string]string{}}
}

func (f *fakeClient) HSet(ctx context.Context, key string, values ...any) IntCmd {
	if f.data[key] == nil {
		f.data[key] = map[string]string{}
	}
	if len(values) == 1 {
		if kvs, ok := values[0].(map[string]any); ok {
			for k, v := range kvs {
				f.data[key][k] = toString(v)
			}
			return fakeIntCmd{}
		}
	}
	for i := 0; i+1 < len(values); i += 2 {
		field, _ := values[i].(string)
		f.data[key][field] = toString(values[i+1])
	}
	return fakeIntCmd{}
}

func (f *fakeClient) HGet(ctx context.Context, key string, field string) StringCmd {
	val, ok := f.data[key][field]
	if !ok {
		return fakeStringCmd{err: errRedisNil{}}
	}
	return fakeStringCmd{val: val}
}

func (f *fakeClient) HDel(ctx context.Context, key string, fields ...string) IntCmd {
	for _, field := range fields {
		delete(f.data[key], field)
	}
	return fakeIntCmd{}
}

func (f *fakeClient) Del(ctx context.Context, keys ...string) IntCmd {
	for _, key := range keys {
		delete(f.data, key)
	}
	return fakeIntCmd{}
}

func (f *fakeClient) Expire(ctx context.Context, key string, expiration time.Duration) BoolCmd {
	return fakeBoolCmd{}
}

type fakeIntCmd struct {
	err error
}

func (c fakeIntCmd) Err() error {
	return c.err
}

type fakeBoolCmd struct {
	err error
}

func (c fakeBoolCmd) Err() error {
	return c.err
}

type fakeStringCmd struct {
	val string
	err error
}

func (c fakeStringCmd) Result() (string, error) {
	return c.val, c.err
}

type errRedisNil struct{}

func (errRedisNil) Error() string {
	return "redis: nil"
}

func toString(val any) string {
	switch v := val.(type) {
	case string:
		return v
	case int64:
		return "123"
	default:
		return ""
	}
}
