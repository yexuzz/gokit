package jwtx

import (
	"context"
	"fmt"
	"time"

	jwtv5 "github.com/golang-jwt/jwt/v5"
)

var _ Manager[struct{}] = (*DefaultManager[struct{}])(nil)

// DefaultManager 是 Manager 的默认实现.
//
// 它负责签发 access token, 签发 refresh token, 校验 token, 检查服务端状态,
// 但不关心 token 来自 HTTP Header, Cookie, gRPC metadata 还是其它地方.
type DefaultManager[T any] struct {
	cfg Config
}

// NewManager 创建默认 JWT Manager.
//
// T 是业务登录数据类型, jwtx 会自动把它放进 JWT 的 data 字段中.
func NewManager[T any](opts ...Option) (*DefaultManager[T], error) {
	cfg := defaultConfig(opts...)
	return &DefaultManager[T]{cfg: cfg}, nil
}

// MustNewManager 创建默认 JWT Manager, 失败时 panic.
//
// 它适合在应用启动初始化阶段使用, 让配置错误尽早暴露.
func MustNewManager[T any](opts ...Option) *DefaultManager[T] {
	manager, err := NewManager[T](opts...)
	if err != nil {
		panic(err)
	}
	return manager
}

// SetLoginToken 登录成功后签发 access token 和 refresh token, 并保存 refresh token jti.
func (h *DefaultManager[T]) SetLoginToken(ctx context.Context, payload T, opts ...IssueOption) (TokenPair, error) {
	issueOpts := issueOptions{}
	for _, opt := range opts {
		opt(&issueOpts)
	}
	if issueOpts.ssid == "" {
		issueOpts.ssid = h.cfg.SSIDGenerator()
	}
	session := Session[T]{
		SSID:      issueOpts.ssid,
		UserAgent: issueOpts.userAgent,
		Payload:   payload,
		Issuer:    h.cfg.Issuer,
	}
	accessToken, err := h.SetAccessToken(ctx, session)
	if err != nil {
		return TokenPair{}, err
	}
	refreshTokenID := h.cfg.TokenIDGenerator()
	now := h.cfg.Now()
	claims := h.newTokenClaims(session, refreshTokenID, now, now.Add(h.cfg.RefreshExpiration))
	refreshToken, err := h.sign(claims, h.cfg.RefreshTokenKey)
	if err != nil {
		return TokenPair{}, err
	}
	if err = h.cfg.Store.SaveRefreshTokenID(ctx, session.SSID, refreshTokenID, h.cfg.RefreshExpiration); err != nil {
		return TokenPair{}, err
	}
	return TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		SSID:         session.SSID,
	}, nil
}

// SetAccessToken 根据已有 session 重新签发 access token.
func (h *DefaultManager[T]) SetAccessToken(ctx context.Context, session Session[T]) (string, error) {
	if session.SSID == "" {
		session.SSID = h.cfg.SSIDGenerator()
	}
	tokenID := h.cfg.TokenIDGenerator()
	now := h.cfg.Now()
	claims := h.newTokenClaims(session, tokenID, now, now.Add(h.cfg.AccessExpiration))
	return h.sign(claims, h.cfg.AccessTokenKey)
}

// CheckAccessToken 校验 access token, 并检查当前 ssid 是否已经被主动失效.
func (h *DefaultManager[T]) CheckAccessToken(ctx context.Context, token string) (Session[T], error) {
	claims, err := h.verify(token, h.cfg.AccessTokenKey)
	if err != nil {
		return Session[T]{}, err
	}
	revoked, err := h.cfg.Store.IsSessionRevoked(ctx, claims.SSID)
	if err != nil {
		return Session[T]{}, err
	}
	if revoked {
		return Session[T]{}, ErrSessionRevoked
	}
	return sessionFromClaims(claims), nil
}

// RefreshAccessToken 校验 refresh token, 确认 refresh token 仍有效后签发新的 access token.
func (h *DefaultManager[T]) RefreshAccessToken(ctx context.Context, refreshToken string) (string, error) {
	claims, err := h.verify(refreshToken, h.cfg.RefreshTokenKey)
	if err != nil {
		return "", err
	}
	revoked, err := h.cfg.Store.IsSessionRevoked(ctx, claims.SSID)
	if err != nil {
		return "", err
	}
	if revoked {
		return "", ErrSessionRevoked
	}
	valid, err := h.cfg.Store.IsRefreshTokenValid(ctx, claims.SSID, claims.ID)
	if err != nil {
		return "", err
	}
	if !valid {
		return "", ErrRefreshTokenInvalid
	}
	return h.SetAccessToken(ctx, sessionFromClaims(claims))
}

// ClearToken 主动失效当前 access token 对应的登录会话.
//
// 它会解析 access token 中的 ssid, 并把 ssid 写入 Store 的失效标记.
func (h *DefaultManager[T]) ClearToken(ctx context.Context, accessToken string) error {
	claims, err := h.verify(accessToken, h.cfg.AccessTokenKey, jwtv5.WithoutClaimsValidation())
	if err != nil {
		return err
	}
	ttl := h.cfg.RefreshExpiration
	if claims.ExpiresAt != nil {
		if remain := time.Until(claims.ExpiresAt.Time); remain > 0 && remain > ttl {
			ttl = remain
		}
	}
	return h.cfg.Store.RevokeSession(ctx, claims.SSID, ttl)
}

// newTokenClaims 根据 session 和有效期创建待签发的 JWT 声明.
func (h *DefaultManager[T]) newTokenClaims(session Session[T], tokenID string, issuedAt time.Time, expiresAt time.Time) *tokenClaims[T] {
	issuer := session.Issuer
	if issuer == "" {
		issuer = h.cfg.Issuer
	}
	return &tokenClaims[T]{
		SSID:      session.SSID,
		UserAgent: session.UserAgent,
		Data:      session.Payload,
		RegisteredClaims: jwtv5.RegisteredClaims{
			ID:        tokenID,
			Issuer:    issuer,
			IssuedAt:  jwtv5.NewNumericDate(issuedAt),
			ExpiresAt: jwtv5.NewNumericDate(expiresAt),
		},
	}
}

// sign 使用指定密钥签发 JWT 字符串.
func (h *DefaultManager[T]) sign(claims *tokenClaims[T], key []byte) (string, error) {
	token := jwtv5.NewWithClaims(h.cfg.SigningMethod, claims)
	return token.SignedString(key)
}

// verify 使用指定密钥校验 JWT 字符串并解析业务登录数据.
func (h *DefaultManager[T]) verify(token string, key []byte, opts ...jwtv5.ParserOption) (*tokenClaims[T], error) {
	if token == "" {
		return nil, ErrInvalidToken
	}
	claims := &tokenClaims[T]{}
	parsed, err := jwtv5.ParseWithClaims(token, claims, func(token *jwtv5.Token) (any, error) {
		if token.Method.Alg() != h.cfg.SigningMethod.Alg() {
			return nil, fmt.Errorf("%w: unexpected signing method %s", ErrInvalidToken, token.Method.Alg())
		}
		return key, nil
	}, opts...)
	if err != nil || parsed == nil || !parsed.Valid {
		return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}
	return claims, nil
}

// sessionFromClaims 将 JWT 声明转换成调用方更容易使用的会话信息.
func sessionFromClaims[T any](claims *tokenClaims[T]) Session[T] {
	session := Session[T]{
		SSID:      claims.SSID,
		TokenID:   claims.ID,
		UserAgent: claims.UserAgent,
		Payload:   claims.Data,
		Issuer:    claims.Issuer,
	}
	if claims.IssuedAt != nil {
		session.IssuedAt = claims.IssuedAt.Time
	}
	if claims.ExpiresAt != nil {
		session.ExpiresAt = claims.ExpiresAt.Time
	}
	return session
}
