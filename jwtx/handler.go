package jwtx

import (
	"context"
	"fmt"
	"time"

	jwtv5 "github.com/golang-jwt/jwt/v5"
)

var _ Handler[*BaseClaims] = (*DefaultHandler[*BaseClaims])(nil)

// DefaultHandler 是 Handler 的默认实现.
//
// 它负责签发 access token, 签发 refresh token, 校验 token, 检查服务端状态,
// 但不关心 token 来自 HTTP Header, Cookie, gRPC metadata 还是其它地方.
type DefaultHandler[T Claims] struct {
	cfg Config[T]
}

// NewHandler 创建默认 JWT Handler.
//
// factory 用于告诉 jwtx 如何创建业务 claims, 例如:
//
//	jwtx.NewHandler(func() *AdminClaims { return &AdminClaims{} }, opts...)
func NewHandler[T Claims](factory ClaimsFactory[T], opts ...Option[T]) (*DefaultHandler[T], error) {
	cfg := defaultConfig(factory, opts...)
	if cfg.ClaimsFactory == nil {
		return nil, ErrClaimsFactoryRequired
	}
	return &DefaultHandler[T]{cfg: cfg}, nil
}

// MustNewHandler 创建默认 JWT Handler, 失败时 panic.
//
// 它适合在应用启动初始化阶段使用, 让配置错误尽早暴露.
func MustNewHandler[T Claims](factory ClaimsFactory[T], opts ...Option[T]) *DefaultHandler[T] {
	handler, err := NewHandler(factory, opts...)
	if err != nil {
		panic(err)
	}
	return handler
}

// SetLoginToken 登录成功后签发 access token 和 refresh token, 并保存 refresh token jti.
func (h *DefaultHandler[T]) SetLoginToken(ctx context.Context, claims T) (TokenPair, error) {
	if claims.GetSSID() == "" {
		claims.SetSSID(h.cfg.SSIDGenerator())
	}
	accessToken, err := h.SetAccessToken(ctx, claims)
	if err != nil {
		return TokenPair{}, err
	}
	refreshTokenID := h.cfg.TokenIDGenerator()
	claims.SetTokenID(refreshTokenID)
	now := h.cfg.Now()
	claims.SetIssuedAt(jwtv5.NewNumericDate(now))
	claims.SetExpiresAt(jwtv5.NewNumericDate(now.Add(h.cfg.RefreshExpiration)))
	claims.SetIssuer(h.cfg.Issuer)
	refreshToken, err := h.sign(claims, h.cfg.RefreshTokenKey)
	if err != nil {
		return TokenPair{}, err
	}
	if err = h.cfg.Store.SaveRefreshTokenID(ctx, claims.GetSSID(), refreshTokenID, h.cfg.RefreshExpiration); err != nil {
		return TokenPair{}, err
	}
	return TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		SSID:         claims.GetSSID(),
	}, nil
}

// SetAccessToken 根据已有 claims 重新签发 access token.
func (h *DefaultHandler[T]) SetAccessToken(ctx context.Context, claims T) (string, error) {
	if claims.GetSSID() == "" {
		claims.SetSSID(h.cfg.SSIDGenerator())
	}
	claims.SetTokenID(h.cfg.TokenIDGenerator())
	now := h.cfg.Now()
	claims.SetIssuedAt(jwtv5.NewNumericDate(now))
	claims.SetExpiresAt(jwtv5.NewNumericDate(now.Add(h.cfg.AccessExpiration)))
	claims.SetIssuer(h.cfg.Issuer)
	return h.sign(claims, h.cfg.AccessTokenKey)
}

// CheckAccessToken 校验 access token, 并检查当前 ssid 是否已经被主动失效.
func (h *DefaultHandler[T]) CheckAccessToken(ctx context.Context, token string) (T, error) {
	claims, err := h.verify(token, h.cfg.AccessTokenKey)
	if err != nil {
		var zero T
		return zero, err
	}
	revoked, err := h.cfg.Store.IsSessionRevoked(ctx, claims.GetSSID())
	if err != nil {
		var zero T
		return zero, err
	}
	if revoked {
		var zero T
		return zero, ErrSessionRevoked
	}
	return claims, nil
}

// RefreshAccessToken 校验 refresh token, 确认 refresh token 仍有效后签发新的 access token.
func (h *DefaultHandler[T]) RefreshAccessToken(ctx context.Context, refreshToken string) (string, error) {
	claims, err := h.verify(refreshToken, h.cfg.RefreshTokenKey)
	if err != nil {
		return "", err
	}
	revoked, err := h.cfg.Store.IsSessionRevoked(ctx, claims.GetSSID())
	if err != nil {
		return "", err
	}
	if revoked {
		return "", ErrSessionRevoked
	}
	valid, err := h.cfg.Store.IsRefreshTokenValid(ctx, claims.GetSSID(), claims.GetTokenID())
	if err != nil {
		return "", err
	}
	if !valid {
		return "", ErrRefreshTokenInvalid
	}
	return h.SetAccessToken(ctx, claims)
}

// ClearToken 主动失效当前 access token 对应的登录会话.
//
// 它会解析 access token 中的 ssid, 并把 ssid 写入 Store 的失效标记.
func (h *DefaultHandler[T]) ClearToken(ctx context.Context, accessToken string) error {
	claims, err := h.verify(accessToken, h.cfg.AccessTokenKey, jwtv5.WithoutClaimsValidation())
	if err != nil {
		return err
	}
	ttl := h.cfg.RefreshExpiration
	if expiresAt, err := claims.GetExpirationTime(); err == nil && expiresAt != nil {
		if remain := time.Until(expiresAt.Time); remain > 0 && remain > ttl {
			ttl = remain
		}
	}
	return h.cfg.Store.RevokeSession(ctx, claims.GetSSID(), ttl)
}

// sign 使用指定密钥签发 JWT 字符串.
func (h *DefaultHandler[T]) sign(claims T, key []byte) (string, error) {
	token := jwtv5.NewWithClaims(h.cfg.SigningMethod, claims)
	return token.SignedString(key)
}

// verify 使用指定密钥校验 JWT 字符串并解析业务 claims.
func (h *DefaultHandler[T]) verify(token string, key []byte, opts ...jwtv5.ParserOption) (T, error) {
	if token == "" {
		var zero T
		return zero, ErrInvalidToken
	}
	claims := h.cfg.ClaimsFactory()
	parsed, err := jwtv5.ParseWithClaims(token, claims, func(token *jwtv5.Token) (any, error) {
		if token.Method.Alg() != h.cfg.SigningMethod.Alg() {
			return nil, fmt.Errorf("%w: unexpected signing method %s", ErrInvalidToken, token.Method.Alg())
		}
		return key, nil
	}, opts...)
	if err != nil || parsed == nil || !parsed.Valid {
		var zero T
		return zero, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}
	return claims, nil
}
