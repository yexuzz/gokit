package jwtx

import (
	"context"
	"time"

	jwtv5 "github.com/golang-jwt/jwt/v5"
)

// Claims 定义 jwtx 需要读写的最小 token 声明能力.
//
// 业务 claims 推荐嵌入 BaseClaims, 然后追加自己的 uid, username, tenant_id 等字段.
type Claims interface {
	jwtv5.Claims

	// GetSSID 返回本次登录会话 ID.
	GetSSID() string
	// SetSSID 写入本次登录会话 ID, 登录签发 token 时会自动生成并写入.
	SetSSID(ssid string)
	// GetTokenID 返回当前 token 的唯一 ID, 对应 JWT 标准字段 jti.
	GetTokenID() string
	// SetTokenID 写入当前 token 的唯一 ID, access token 和 refresh token 会使用不同 jti.
	SetTokenID(tokenID string)
	// SetIssuedAt 写入 JWT 签发时间.
	SetIssuedAt(issuedAt *jwtv5.NumericDate)
	// SetExpiresAt 写入 JWT 过期时间.
	SetExpiresAt(expiresAt *jwtv5.NumericDate)
	// SetIssuer 写入 JWT 签发方.
	SetIssuer(issuer string)
}

// BaseClaims 是业务 claims 可以嵌入的 JWT 基础声明.
//
// 它包含 jwtx 统一处理登录态所需的 ssid, user_agent 和 jwt.RegisteredClaims.
// 嵌入后 JSON 仍然是扁平结构, 不会额外出现 base_claims 层级.
type BaseClaims struct {
	SSID      string `json:"ssid"`                 // 本次登录会话 ID.
	UserAgent string `json:"user_agent,omitempty"` // 登录或请求时的 User-Agent, 可由业务自行校验.
	jwtv5.RegisteredClaims
}

// GetSSID 返回本次登录会话 ID.
func (c *BaseClaims) GetSSID() string {
	if c == nil {
		return ""
	}
	return c.SSID
}

// SetSSID 写入本次登录会话 ID.
func (c *BaseClaims) SetSSID(ssid string) {
	if c != nil {
		c.SSID = ssid
	}
}

// GetTokenID 返回 JWT 标准字段 jti.
func (c *BaseClaims) GetTokenID() string {
	if c == nil {
		return ""
	}
	return c.ID
}

// SetTokenID 写入 JWT 标准字段 jti.
func (c *BaseClaims) SetTokenID(tokenID string) {
	if c != nil {
		c.ID = tokenID
	}
}

// SetIssuedAt 写入 JWT 签发时间 iat.
func (c *BaseClaims) SetIssuedAt(issuedAt *jwtv5.NumericDate) {
	if c != nil {
		c.IssuedAt = issuedAt
	}
}

// SetExpiresAt 写入 JWT 过期时间 exp.
func (c *BaseClaims) SetExpiresAt(expiresAt *jwtv5.NumericDate) {
	if c != nil {
		c.ExpiresAt = expiresAt
	}
}

// SetIssuer 写入 JWT 签发方 iss.
func (c *BaseClaims) SetIssuer(issuer string) {
	if c != nil {
		c.Issuer = issuer
	}
}

// TokenPair 表示一次登录签发出来的长短 token.
type TokenPair struct {
	AccessToken  string // 短期 access token, 用于普通接口鉴权.
	RefreshToken string // 长期 refresh token, 用于刷新 access token.
	SSID         string // 本次登录会话 ID, 方便业务记录登录日志.
}

// Handler 定义 JWT 登录凭证的签发, 校验, 刷新和清理流程.
//
// Handler 不依赖 Gin, Cookie 或 Redis 具体类型. HTTP 框架只需要负责取出 token 字符串,
// 再把返回的 token 写回响应即可.
type Handler[T Claims] interface {
	// SetLoginToken 登录成功后签发 access token 和 refresh token, 并保存 refresh token 状态.
	SetLoginToken(ctx context.Context, claims T) (TokenPair, error)
	// SetAccessToken 根据已有 claims 重新签发 access token.
	SetAccessToken(ctx context.Context, claims T) (string, error)
	// CheckAccessToken 校验 access token, 并检查当前 ssid 是否已经被主动失效.
	CheckAccessToken(ctx context.Context, token string) (T, error)
	// RefreshAccessToken 校验 refresh token, 确认 refresh token 仍有效后签发新的 access token.
	RefreshAccessToken(ctx context.Context, refreshToken string) (string, error)
	// ClearToken 主动失效当前 access token 对应的登录会话.
	ClearToken(ctx context.Context, accessToken string) error
}

// Store 定义 jwtx 需要的服务端 token 状态存储能力.
//
// 它不是传统 session 存储, 只保存两类状态:
//  1. 某个 ssid 是否已经退出或被主动失效.
//  2. 某个 ssid 当前有效的 refresh token jti 是哪个.
type Store interface {
	// RevokeSession 标记一个登录会话已经失效, 后续 access/refresh token 都不能再使用.
	RevokeSession(ctx context.Context, ssid string, ttl time.Duration) error
	// IsSessionRevoked 检查登录会话是否已经被主动失效.
	IsSessionRevoked(ctx context.Context, ssid string) (bool, error)
	// SaveRefreshTokenID 保存当前登录会话有效的 refresh token jti.
	SaveRefreshTokenID(ctx context.Context, ssid string, tokenID string, ttl time.Duration) error
	// IsRefreshTokenValid 检查传入的 refresh token jti 是否仍是当前登录会话的有效 jti.
	IsRefreshTokenValid(ctx context.Context, ssid string, tokenID string) (bool, error)
}

// ClaimsFactory 创建用于解析 JWT 的空 claims.
//
// jwt.ParseWithClaims 需要调用方提供具体 claims 类型, 泛型场景下由业务传入这个工厂函数.
type ClaimsFactory[T Claims] func() T
