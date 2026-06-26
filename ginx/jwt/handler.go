package jwt

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/yexuzz/gokit/ginx"
	"github.com/yexuzz/gokit/jwtx"
)

const (
	// DefaultAccessTokenHeader 表示响应中写入 access token 的默认 header.
	DefaultAccessTokenHeader = "x-jwt-token"
	// DefaultRefreshTokenHeader 表示响应中写入 refresh token 的默认 header.
	DefaultRefreshTokenHeader = "x-refresh-token"
)

var (
	// ErrUserAgentMismatch 表示 token 中记录的 User-Agent 和当前请求不一致.
	ErrUserAgentMismatch = errors.New("ginx/jwt: user agent mismatch")
)

// TokenExtractor 定义从 Gin 请求中提取 token 字符串的函数.
type TokenExtractor func(ctx *gin.Context) string

// ErrorHandler 定义登录态校验失败时的响应处理函数.
type ErrorHandler func(ctx *gin.Context, err error)

// Option 用于调整 Gin JWT 适配层配置.
type Option func(*Config)

// Config 定义 Gin JWT 适配层的运行配置.
type Config struct {
	AccessTokenHeader     string         // 写入 access token 的响应 header.
	RefreshTokenHeader    string         // 写入 refresh token 的响应 header.
	AccessTokenExtractor  TokenExtractor // 从请求中提取 access token 的函数.
	RefreshTokenExtractor TokenExtractor // 从请求中提取 refresh token 的函数.
	ErrorHandler          ErrorHandler   // 中间件校验失败时的响应处理函数.
	CheckUserAgent        bool           // 是否校验 token 中的 User-Agent.
}

// WithAccessTokenHeader 设置响应中写入 access token 的 header 名称.
func WithAccessTokenHeader(header string) Option {
	return func(cfg *Config) {
		cfg.AccessTokenHeader = header
	}
}

// WithRefreshTokenHeader 设置响应中写入 refresh token 的 header 名称.
func WithRefreshTokenHeader(header string) Option {
	return func(cfg *Config) {
		cfg.RefreshTokenHeader = header
	}
}

// WithAccessTokenExtractor 设置从请求中提取 access token 的函数.
func WithAccessTokenExtractor(extractor TokenExtractor) Option {
	return func(cfg *Config) {
		cfg.AccessTokenExtractor = extractor
	}
}

// WithRefreshTokenExtractor 设置从请求中提取 refresh token 的函数.
func WithRefreshTokenExtractor(extractor TokenExtractor) Option {
	return func(cfg *Config) {
		cfg.RefreshTokenExtractor = extractor
	}
}

// WithErrorHandler 设置登录态校验失败时的响应处理函数.
func WithErrorHandler(handler ErrorHandler) Option {
	return func(cfg *Config) {
		cfg.ErrorHandler = handler
	}
}

// WithUserAgentCheck 设置是否校验 token 中记录的 User-Agent.
func WithUserAgentCheck(check bool) Option {
	return func(cfg *Config) {
		cfg.CheckUserAgent = check
	}
}

// Handler 把通用 jwtx.Manager 适配成 Gin 项目里更顺手的登录态组件.
//
// 它负责从请求中提取 token, 把校验后的 session 写入 gin.Context,
// 并在登录和刷新时把新 token 写入响应 header.
type Handler[T any] struct {
	manager jwtx.Manager[T]
	cfg     Config
}

// NewHandler 创建 Gin JWT 适配层.
func NewHandler[T any](manager jwtx.Manager[T], opts ...Option) *Handler[T] {
	cfg := Config{
		AccessTokenHeader:     DefaultAccessTokenHeader,
		RefreshTokenHeader:    DefaultRefreshTokenHeader,
		AccessTokenExtractor:  ExtractBearerToken,
		RefreshTokenExtractor: ExtractBearerToken,
		ErrorHandler:          defaultErrorHandler,
		CheckUserAgent:        true,
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	if cfg.AccessTokenExtractor == nil {
		cfg.AccessTokenExtractor = ExtractBearerToken
	}
	if cfg.RefreshTokenExtractor == nil {
		cfg.RefreshTokenExtractor = ExtractBearerToken
	}
	return &Handler[T]{
		manager: manager,
		cfg:     cfg,
	}
}

// ExtractBearerToken 从 Authorization header 中提取 Bearer token.
//
// 如果 header 不是 Bearer 格式, 会把完整 header 当成 token 返回, 兼容老项目直接传 token 的写法.
func ExtractBearerToken(ctx *gin.Context) string {
	value := strings.TrimSpace(ctx.GetHeader("Authorization"))
	if value == "" {
		return ""
	}
	prefix := "Bearer "
	if len(value) >= len(prefix) && strings.EqualFold(value[:len(prefix)], prefix) {
		return strings.TrimSpace(value[len(prefix):])
	}
	return value
}

// ExtractHeader 从指定 header 中提取 token 字符串.
func ExtractHeader(header string) TokenExtractor {
	return func(ctx *gin.Context) string {
		return strings.TrimSpace(ctx.GetHeader(header))
	}
}

// SetLoginToken 在登录成功后签发 access token 和 refresh token, 并写入响应 header.
func (h *Handler[T]) SetLoginToken(ctx *gin.Context, payload T, opts ...jwtx.IssueOption) (string, error) {
	issueOpts := []jwtx.IssueOption{jwtx.WithUserAgent(ctx.GetHeader("User-Agent"))}
	issueOpts = append(issueOpts, opts...)
	pair, err := h.manager.SetLoginToken(ctx.Request.Context(), payload, issueOpts...)
	if err != nil {
		return "", err
	}
	ctx.Header(h.cfg.AccessTokenHeader, pair.AccessToken)
	ctx.Header(h.cfg.RefreshTokenHeader, pair.RefreshToken)
	return pair.SSID, nil
}

// SetAccessToken 根据已有 session 重新签发 access token, 并写入响应 header.
func (h *Handler[T]) SetAccessToken(ctx *gin.Context, session jwtx.Session[T]) (jwtx.AccessToken, error) {
	accessToken, err := h.manager.SetAccessToken(ctx, session)
	if err != nil {
		return jwtx.AccessToken{}, err
	}
	ctx.Header(h.cfg.AccessTokenHeader, accessToken.Token)
	return accessToken, nil
}

// RefreshToken 从当前请求中读取 refresh token, 校验通过后写入新的 access token.
func (h *Handler[T]) RefreshToken(ctx *gin.Context) (jwtx.AccessToken, error) {
	refreshToken := h.cfg.RefreshTokenExtractor(ctx)
	accessToken, err := h.manager.RefreshAccessToken(ctx.Request.Context(), refreshToken)
	if err != nil {
		return jwtx.AccessToken{}, err
	}
	ctx.Header(h.cfg.AccessTokenHeader, accessToken.Token)
	return accessToken, nil
}

// CheckToken 校验当前请求中的 access token, 并返回登录会话.
func (h *Handler[T]) CheckToken(ctx *gin.Context) (jwtx.Session[T], error) {
	token := h.cfg.AccessTokenExtractor(ctx)
	session, err := h.manager.CheckAccessToken(ctx.Request.Context(), token)
	if err != nil {
		return jwtx.Session[T]{}, err
	}
	if h.cfg.CheckUserAgent && session.UserAgent != "" && session.UserAgent != ctx.GetHeader("User-Agent") {
		return jwtx.Session[T]{}, ErrUserAgentMismatch
	}
	return session, nil
}

// ClearToken 清除当前登录会话的 token, 并把当前 ssid 标记为失效.
//
// 如果中间件已经把 session 写入 gin.Context, 会优先按 ssid 清理;
// 否则会回退为解析当前请求里的 access token 后清理.
func (h *Handler[T]) ClearToken(ctx *gin.Context) error {
	ctx.Header(h.cfg.AccessTokenHeader, "")
	ctx.Header(h.cfg.RefreshTokenHeader, "")
	if session, ok := Session[T](ctx); ok {
		return h.manager.ClearUserSession(ctx.Request.Context(), session.UserID, session.SSID)
	}
	return h.manager.ClearToken(ctx.Request.Context(), h.cfg.AccessTokenExtractor(ctx))
}

// ClearUserSessions 清除当前登录用户的全部设备会话.
func (h *Handler[T]) ClearUserSessions(ctx *gin.Context) error {
	session, ok := Session[T](ctx)
	if !ok {
		return nil
	}
	return h.ClearUserSessionsByUserID(ctx, session.UserID)
}

// ClearUserSessionsByUserID 清除指定用户的全部设备会话.
func (h *Handler[T]) ClearUserSessionsByUserID(ctx *gin.Context, userID string) error {
	ctx.Header(h.cfg.AccessTokenHeader, "")
	ctx.Header(h.cfg.RefreshTokenHeader, "")
	return h.manager.ClearUserSessions(ctx.Request.Context(), userID)
}

// defaultErrorHandler 返回统一的未登录响应.
func defaultErrorHandler(ctx *gin.Context, err error) {
	ctx.AbortWithStatusJSON(http.StatusUnauthorized, ginx.NewResult(http.StatusUnauthorized, "未登录"))
}
