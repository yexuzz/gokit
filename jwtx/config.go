package jwtx

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	jwtv5 "github.com/golang-jwt/jwt/v5"
)

// Config 定义 JWT Manager 的运行配置.
type Config struct {
	AccessTokenKey    []byte                   // access token 签名密钥.
	RefreshTokenKey   []byte                   // refresh token 签名密钥.
	AccessExpiration  time.Duration            // access token 有效期.
	RefreshExpiration time.Duration            // refresh token 有效期.
	SigningMethod     jwtv5.SigningMethod      // JWT 签名算法, 默认 HS256.
	Issuer            string                   // JWT 签发方.
	Store             Store                    // token 状态存储, 默认 NoopStore.
	UserIDExtractor   func(payload any) string // 从业务 payload 中提取用户 ID, 用于维护用户多设备会话集合.
	Now               func() time.Time         // 当前时间函数, 主要用于测试.
	SSIDGenerator     func() string            // ssid 生成函数.
	TokenIDGenerator  func() string            // jti 生成函数.
}

// Option 用于调整 JWT Manager 配置.
type Option func(*Config)

// WithAccessTokenKey 设置 access token 签名密钥.
func WithAccessTokenKey(key []byte) Option {
	return func(cfg *Config) {
		cfg.AccessTokenKey = key
	}
}

// WithRefreshTokenKey 设置 refresh token 签名密钥.
func WithRefreshTokenKey(key []byte) Option {
	return func(cfg *Config) {
		cfg.RefreshTokenKey = key
	}
}

// WithExpiration 设置长短 token 的有效期.
func WithExpiration(accessExpiration time.Duration, refreshExpiration time.Duration) Option {
	return func(cfg *Config) {
		cfg.AccessExpiration = accessExpiration
		cfg.RefreshExpiration = refreshExpiration
	}
}

// WithSigningMethod 设置 JWT 签名算法.
func WithSigningMethod(method jwtv5.SigningMethod) Option {
	return func(cfg *Config) {
		cfg.SigningMethod = method
	}
}

// WithIssuer 设置 JWT 签发方.
func WithIssuer(issuer string) Option {
	return func(cfg *Config) {
		cfg.Issuer = issuer
	}
}

// WithStore 设置 token 状态存储.
func WithStore(store Store) Option {
	return func(cfg *Config) {
		cfg.Store = store
	}
}

// WithUserIDExtractor 设置从业务 payload 中提取用户 ID 的函数.
//
// 配置后, jwtx 会在登录成功时维护 userID -> ssid 集合, 从而支持退出指定用户的全部设备.
func WithUserIDExtractor[T any](extractor func(payload T) string) Option {
	return func(cfg *Config) {
		if extractor == nil {
			cfg.UserIDExtractor = nil
			return
		}
		cfg.UserIDExtractor = func(payload any) string {
			val, ok := payload.(T)
			if !ok {
				return ""
			}
			return extractor(val)
		}
	}
}

// WithNow 设置当前时间函数.
func WithNow(now func() time.Time) Option {
	return func(cfg *Config) {
		cfg.Now = now
	}
}

// WithSSIDGenerator 设置 ssid 生成函数.
func WithSSIDGenerator(fn func() string) Option {
	return func(cfg *Config) {
		cfg.SSIDGenerator = fn
	}
}

// WithTokenIDGenerator 设置 JWT jti 生成函数.
func WithTokenIDGenerator(fn func() string) Option {
	return func(cfg *Config) {
		cfg.TokenIDGenerator = fn
	}
}

// defaultConfig 构造默认配置并应用调用方传入的 Option.
func defaultConfig(opts ...Option) Config {
	cfg := Config{
		AccessExpiration:  15 * time.Minute,
		RefreshExpiration: 7 * 24 * time.Hour,
		SigningMethod:     jwtv5.SigningMethodHS256,
		Store:             NoopStore{},
		Now:               time.Now,
		SSIDGenerator:     randomID,
		TokenIDGenerator:  randomID,
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	if cfg.RefreshTokenKey == nil {
		cfg.RefreshTokenKey = cfg.AccessTokenKey
	}
	if cfg.SigningMethod == nil {
		cfg.SigningMethod = jwtv5.SigningMethodHS256
	}
	if cfg.Store == nil {
		cfg.Store = NoopStore{}
	}
	if cfg.Now == nil {
		cfg.Now = time.Now
	}
	if cfg.SSIDGenerator == nil {
		cfg.SSIDGenerator = randomID
	}
	if cfg.TokenIDGenerator == nil {
		cfg.TokenIDGenerator = randomID
	}
	return cfg
}

// randomID 生成默认 ssid 和 jti.
func randomID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return hex.EncodeToString([]byte(time.Now().Format(time.RFC3339Nano)))
	}
	return hex.EncodeToString(b[:])
}
