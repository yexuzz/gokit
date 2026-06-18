package session

import (
	"fmt"
	"time"

	jwtv5 "github.com/golang-jwt/jwt/v5"
)

// TokenManager 定义 access token 的签发和校验能力。
//
// 这里抽成接口，是为了让 Provider 不绑定某一种 JWT 实现。
// 默认实现使用 github.com/golang-jwt/jwt/v5。
type TokenManager interface {
	// GenerateAccessToken 根据 Claims 签发 access token。
	GenerateAccessToken(claims Claims) (string, error)

	// VerifyAccessToken 校验 access token 并解析 Claims。
	VerifyAccessToken(token string) (Claims, error)
}

// JWTConfig 定义默认 JWT 管理器的配置。
type JWTConfig struct {
	// Expiration 表示 access token 的有效期。
	Expiration time.Duration

	// SigningKey 表示签名密钥。
	SigningKey string

	// VerifyKey 表示验签密钥。
	//
	// 对 HS256 这类对称算法来说，VerifyKey 通常和 SigningKey 相同。
	VerifyKey string

	// SigningMethod 表示 JWT 签名算法，默认使用 HS256。
	SigningMethod jwtv5.SigningMethod

	// Issuer 表示 JWT 签发方。
	Issuer string

	// Now 返回当前时间。
	//
	// 测试时可以替换成固定时间，生产环境一般保持 nil，内部会使用 time.Now。
	Now func() time.Time

	// GenID 生成 JWT ID，也就是 RegisteredClaims.ID。
	GenID func() string
}

// JWTOption 用于调整 JWTConfig。
type JWTOption func(*JWTConfig)

// WithJWTVerifyKey 设置 JWT 验签密钥。
func WithJWTVerifyKey(key string) JWTOption {
	return func(cfg *JWTConfig) {
		cfg.VerifyKey = key
	}
}

// WithJWTSigningMethod 设置 JWT 签名算法。
func WithJWTSigningMethod(method jwtv5.SigningMethod) JWTOption {
	return func(cfg *JWTConfig) {
		cfg.SigningMethod = method
	}
}

// WithJWTIssuer 设置 JWT 签发方。
func WithJWTIssuer(issuer string) JWTOption {
	return func(cfg *JWTConfig) {
		cfg.Issuer = issuer
	}
}

// WithJWTNow 设置 JWT 管理器使用的当前时间函数。
func WithJWTNow(now func() time.Time) JWTOption {
	return func(cfg *JWTConfig) {
		cfg.Now = now
	}
}

// WithJWTGenID 设置 JWT ID 生成函数。
func WithJWTGenID(fn func() string) JWTOption {
	return func(cfg *JWTConfig) {
		cfg.GenID = fn
	}
}

// JWTManager 是基于 golang-jwt/jwt/v5 的 TokenManager 实现。
type JWTManager struct {
	cfg JWTConfig
}

// NewJWTManager 创建默认 JWT 管理器。
func NewJWTManager(expiration time.Duration, signingKey string, opts ...JWTOption) *JWTManager {
	cfg := JWTConfig{
		Expiration:    expiration,
		SigningKey:    signingKey,
		VerifyKey:     signingKey,
		SigningMethod: jwtv5.SigningMethodHS256,
		Now:           time.Now,
		GenID:         func() string { return "" },
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	if cfg.Now == nil {
		cfg.Now = time.Now
	}
	if cfg.GenID == nil {
		cfg.GenID = func() string { return "" }
	}
	if cfg.VerifyKey == "" {
		cfg.VerifyKey = cfg.SigningKey
	}
	return &JWTManager{cfg: cfg}
}

// GenerateAccessToken 根据 Claims 签发 access token。
func (m *JWTManager) GenerateAccessToken(claims Claims) (string, error) {
	now := m.cfg.Now()
	registered := registeredClaims{
		Data: claims,
		RegisteredClaims: jwtv5.RegisteredClaims{
			Issuer:    m.cfg.Issuer,
			ExpiresAt: jwtv5.NewNumericDate(now.Add(m.cfg.Expiration)),
			IssuedAt:  jwtv5.NewNumericDate(now),
			ID:        m.cfg.GenID(),
		},
	}
	token := jwtv5.NewWithClaims(m.cfg.SigningMethod, registered)
	return token.SignedString([]byte(m.cfg.SigningKey))
}

// VerifyAccessToken 校验 access token 并解析 Claims。
func (m *JWTManager) VerifyAccessToken(token string) (Claims, error) {
	parsed, err := jwtv5.ParseWithClaims(token, &registeredClaims{}, func(*jwtv5.Token) (any, error) {
		return []byte(m.cfg.VerifyKey), nil
	})
	if err != nil || !parsed.Valid {
		return Claims{}, fmt.Errorf("%w: token 校验失败: %v", ErrUnauthorized, err)
	}
	claims, ok := parsed.Claims.(*registeredClaims)
	if !ok {
		return Claims{}, fmt.Errorf("%w: token claims 类型错误", ErrUnauthorized)
	}
	return claims.Data, nil
}

type registeredClaims struct {
	Data Claims `json:"data"`
	jwtv5.RegisteredClaims
}
