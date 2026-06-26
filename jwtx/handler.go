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
		UserID:    h.userID(payload),
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
	if session.UserID != "" {
		if store, ok := h.cfg.Store.(UserSessionStore); ok {
			if err = store.AddUserSession(ctx, session.UserID, session.SSID, h.cfg.RefreshExpiration); err != nil {
				return TokenPair{}, err
			}
		}
	}
	return TokenPair{
		AccessToken:  accessToken.Token,
		RefreshToken: refreshToken,
		SSID:         accessToken.SSID,
	}, nil
}

// SetAccessToken 根据已有 session 重新签发 access token, 并返回本次签发的 ssid 和 jti.
func (h *DefaultManager[T]) SetAccessToken(ctx context.Context, session Session[T]) (AccessToken, error) {
	if session.SSID == "" {
		session.SSID = h.cfg.SSIDGenerator()
	}
	if session.UserID == "" {
		session.UserID = h.userID(session.Payload)
	}
	tokenID := h.cfg.TokenIDGenerator()
	now := h.cfg.Now()
	expiresAt := now.Add(h.cfg.AccessExpiration)
	claims := h.newTokenClaims(session, tokenID, now, expiresAt)
	token, err := h.sign(claims, h.cfg.AccessTokenKey)
	if err != nil {
		return AccessToken{}, err
	}
	return AccessToken{
		Token:     token,
		SSID:      session.SSID,
		TokenID:   tokenID,
		ExpiresAt: expiresAt,
	}, nil
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
func (h *DefaultManager[T]) RefreshAccessToken(ctx context.Context, refreshToken string) (AccessToken, error) {
	claims, err := h.verify(refreshToken, h.cfg.RefreshTokenKey)
	if err != nil {
		return AccessToken{}, err
	}
	revoked, err := h.cfg.Store.IsSessionRevoked(ctx, claims.SSID)
	if err != nil {
		return AccessToken{}, err
	}
	if revoked {
		return AccessToken{}, ErrSessionRevoked
	}
	valid, err := h.cfg.Store.IsRefreshTokenValid(ctx, claims.SSID, claims.ID)
	if err != nil {
		return AccessToken{}, err
	}
	if !valid {
		return AccessToken{}, ErrRefreshTokenInvalid
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
	if claims.SSID == "" {
		return ErrInvalidToken
	}
	ttl := h.cfg.RefreshExpiration
	if claims.ExpiresAt != nil {
		if remain := time.Until(claims.ExpiresAt.Time); remain > 0 && remain > ttl {
			ttl = remain
		}
	}
	return h.clearSession(ctx, claims.UserID, claims.SSID, ttl)
}

// ClearSession 按 ssid 主动失效登录会话, 后续同会话 access/refresh token 都不能再使用.
func (h *DefaultManager[T]) ClearSession(ctx context.Context, ssid string) error {
	if ssid == "" {
		return ErrInvalidToken
	}
	return h.cfg.Store.RevokeSession(ctx, ssid, h.cfg.RefreshExpiration)
}

// ClearUserSession 按 userID 和 ssid 主动失效一个登录会话, 并从用户会话集合中移除.
func (h *DefaultManager[T]) ClearUserSession(ctx context.Context, userID string, ssid string) error {
	if ssid == "" {
		return ErrInvalidToken
	}
	return h.clearSession(ctx, userID, ssid, h.cfg.RefreshExpiration)
}

// ClearUserSessions 主动失效指定用户的全部登录会话.
func (h *DefaultManager[T]) ClearUserSessions(ctx context.Context, userID string) error {
	store, ok := h.cfg.Store.(UserSessionStore)
	if !ok {
		return ErrUserSessionStoreUnsupported
	}

	ssids, err := store.ListUserSessions(ctx, userID)
	if err != nil {
		return err
	}
	for _, ssid := range ssids {
		if ssid == "" {
			continue
		}
		if err = h.cfg.Store.RevokeSession(ctx, ssid, h.cfg.RefreshExpiration); err != nil {
			return err
		}
	}
	return store.ClearUserSessions(ctx, userID)
}

// clearSession 主动失效一个登录会话, 并在 userID 存在时维护用户会话集合.
func (h *DefaultManager[T]) clearSession(ctx context.Context, userID string, ssid string, ttl time.Duration) error {
	if err := h.cfg.Store.RevokeSession(ctx, ssid, ttl); err != nil {
		return err
	}
	if userID == "" {
		return nil
	}
	if store, ok := h.cfg.Store.(UserSessionStore); ok {
		return store.RemoveUserSession(ctx, userID, ssid)
	}
	return nil
}

// newTokenClaims 根据 session 和有效期创建待签发的 JWT 声明.
func (h *DefaultManager[T]) newTokenClaims(session Session[T], tokenID string, issuedAt time.Time, expiresAt time.Time) *tokenClaims[T] {
	issuer := session.Issuer
	if issuer == "" {
		issuer = h.cfg.Issuer
	}
	return &tokenClaims[T]{
		SSID:      session.SSID,
		UserID:    session.UserID,
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
		UserID:    claims.UserID,
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

// userID 从业务 payload 中提取用户 ID, 未配置时返回空字符串.
func (h *DefaultManager[T]) userID(payload T) string {
	if h.cfg.UserIDExtractor == nil {
		return ""
	}
	return h.cfg.UserIDExtractor(payload)
}
