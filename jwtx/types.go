package jwtx

import (
	"context"
	"time"

	jwtv5 "github.com/golang-jwt/jwt/v5"
)

// tokenClaims 表示 jwtx 写入 JWT 的完整声明.
//
// Data 保存业务自己的登录数据, ssid, jti, exp, iat 等 token 元信息由 jwtx 维护.
type tokenClaims[T any] struct {
	SSID      string `json:"ssid"`                 // 本次登录会话 ID.
	UserAgent string `json:"user_agent,omitempty"` // 登录或请求时的 User-Agent, 可由业务自行校验.
	Data      T      `json:"data"`                 // 业务登录数据.
	jwtv5.RegisteredClaims
}

// Session 表示校验 token 后得到的登录会话信息.
type Session[T any] struct {
	SSID      string    // 本次登录会话 ID.
	TokenID   string    // 当前 token 的唯一 ID, 对应 JWT 标准字段 jti.
	UserAgent string    // 登录或请求时的 User-Agent.
	Payload   T         // 业务登录数据.
	Issuer    string    // JWT 签发方.
	IssuedAt  time.Time // JWT 签发时间.
	ExpiresAt time.Time // JWT 过期时间.
}

// TokenPair 表示一次登录签发出来的长短 token.
type TokenPair struct {
	AccessToken  string // 短期 access token, 用于普通接口鉴权.
	RefreshToken string // 长期 refresh token, 用于刷新 access token.
	SSID         string // 本次登录会话 ID, 方便业务记录登录日志.
}

// IssueOption 表示签发 token 时可选的会话参数.
type IssueOption func(*issueOptions)

type issueOptions struct {
	ssid      string
	userAgent string
}

// WithSessionID 指定本次登录会话 ID, 不传时由 jwtx 自动生成.
func WithSessionID(ssid string) IssueOption {
	return func(opts *issueOptions) {
		opts.ssid = ssid
	}
}

// WithUserAgent 指定本次登录或请求时的 User-Agent.
func WithUserAgent(userAgent string) IssueOption {
	return func(opts *issueOptions) {
		opts.userAgent = userAgent
	}
}

// Handler 定义 JWT 登录凭证的签发, 校验, 刷新和清理流程.
//
// Handler 不依赖 Gin, Cookie 或 Redis 具体类型. HTTP 框架只需要负责取出 token 字符串,
// 再把返回的 token 写回响应即可.
type Handler[T any] interface {
	// SetLoginToken 登录成功后签发 access token 和 refresh token, 并保存 refresh token 状态.
	SetLoginToken(ctx context.Context, payload T, opts ...IssueOption) (TokenPair, error)
	// SetAccessToken 根据已有 session 重新签发 access token.
	SetAccessToken(ctx context.Context, session Session[T]) (string, error)
	// CheckAccessToken 校验 access token, 并检查当前 ssid 是否已经被主动失效.
	CheckAccessToken(ctx context.Context, token string) (Session[T], error)
	// RefreshAccessToken 校验 refresh token, 确认 refresh token 仍有效后签发新的 access token.
	RefreshAccessToken(ctx context.Context, refreshToken string) (string, error)
	// ClearToken 主动失效当前 access token 对应的登录会话.
	ClearToken(ctx context.Context, accessToken string) error
}

// Store 定义 jwtx 需要的服务端 token 状态存储能力.
//
// 它不是传统 session 存储, 只保存两类状态:
//  1. 某个 ssid 是否已经退出或被主动失效.
//  2. 某个 ssid 当前有效的 refresh token jti 是哪一个.
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
