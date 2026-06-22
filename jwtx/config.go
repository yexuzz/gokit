package jwtx

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	jwtv5 "github.com/golang-jwt/jwt/v5"
)

// Config 定义 JWT Handler 的运行配置.
type Config[T Claims] struct {
	AccessTokenKey    []byte              // access token 签名密钥.
	RefreshTokenKey   []byte              // refresh token 签名密钥.
	AccessExpiration  time.Duration       // access token 有效期.
	RefreshExpiration time.Duration       // refresh token 有效期.
	SigningMethod     jwtv5.SigningMethod // JWT 签名算法, 默认 HS256.
	Issuer            string              // JWT 签发方.
	Store             Store               // token 状态存储, 默认 NoopStore.
	ClaimsFactory     ClaimsFactory[T]    // claims 工厂函数, 用于解析 token.
	Now               func() time.Time    // 当前时间函数, 主要用于测试.
	SSIDGenerator     func() string       // ssid 生成函数.
	TokenIDGenerator  func() string       // jti 生成函数.
}

// Option 用于调整 JWT Handler 配置.
type Option[T Claims] func(*Config[T])

// WithAccessTokenKey 设置 access token 签名密钥.
func WithAccessTokenKey[T Claims](key []byte) Option[T] {
	return func(cfg *Config[T]) {
		cfg.AccessTokenKey = key
	}
}

// WithRefreshTokenKey 设置 refresh token 签名密钥.
func WithRefreshTokenKey[T Claims](key []byte) Option[T] {
	return func(cfg *Config[T]) {
		cfg.RefreshTokenKey = key
	}
}

// WithExpiration 设置长短 token 的有效期.
func WithExpiration[T Claims](accessExpiration time.Duration, refreshExpiration time.Duration) Option[T] {
	return func(cfg *Config[T]) {
		cfg.AccessExpiration = accessExpiration
		cfg.RefreshExpiration = refreshExpiration
	}
}

// WithSigningMethod 设置 JWT 签名算法.
func WithSigningMethod[T Claims](method jwtv5.SigningMethod) Option[T] {
	return func(cfg *Config[T]) {
		cfg.SigningMethod = method
	}
}

// WithIssuer 设置 JWT 签发方.
func WithIssuer[T Claims](issuer string) Option[T] {
	return func(cfg *Config[T]) {
		cfg.Issuer = issuer
	}
}

// WithStore 设置 token 状态存储.
func WithStore[T Claims](store Store) Option[T] {
	return func(cfg *Config[T]) {
		cfg.Store = store
	}
}

// WithNow 设置当前时间函数.
func WithNow[T Claims](now func() time.Time) Option[T] {
	return func(cfg *Config[T]) {
		cfg.Now = now
	}
}

// WithSSIDGenerator 设置 ssid 生成函数.
func WithSSIDGenerator[T Claims](fn func() string) Option[T] {
	return func(cfg *Config[T]) {
		cfg.SSIDGenerator = fn
	}
}

// WithTokenIDGenerator 设置 JWT jti 生成函数.
func WithTokenIDGenerator[T Claims](fn func() string) Option[T] {
	return func(cfg *Config[T]) {
		cfg.TokenIDGenerator = fn
	}
}

// defaultConfig 构造默认配置并应用调用方传入的 Option.
func defaultConfig[T Claims](factory ClaimsFactory[T], opts ...Option[T]) Config[T] {
	cfg := Config[T]{
		AccessExpiration:  15 * time.Minute,
		RefreshExpiration: 7 * 24 * time.Hour,
		SigningMethod:     jwtv5.SigningMethodHS256,
		Store:             NoopStore{},
		ClaimsFactory:     factory,
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
