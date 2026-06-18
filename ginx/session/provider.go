package session

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

var _ Provider = (*SessionProvider)(nil)

// SessionProvider 是默认的 Provider 实现。
//
// 它把 JWT、TokenCarrier 和 Store 组合在一起：
//  1. JWT 负责证明“这个请求是谁，以及它对应哪个 ssid”。
//  2. TokenCarrier 负责决定 token 放在哪里，例如响应头或 Cookie。
//  3. Store 负责保存真正的服务端 Session 数据。
type SessionProvider struct {
	store        Store
	manager      TokenManager
	carrier      TokenCarrier
	expiration   time.Duration
	generateSSID func() string
}

// ProviderOption 用于调整 SessionProvider。
type ProviderOption func(*SessionProvider)

// WithTokenCarrier 设置 token 的输入输出方式。
func WithTokenCarrier(carrier TokenCarrier) ProviderOption {
	return func(p *SessionProvider) {
		p.carrier = carrier
	}
}

// WithTokenManager 设置 token 管理器。
func WithTokenManager(manager TokenManager) ProviderOption {
	return func(p *SessionProvider) {
		p.manager = manager
	}
}

// WithSSIDGenerator 设置 ssid 生成函数。
//
// 测试时可以传入固定值，生产环境通常保持默认的随机生成。
func WithSSIDGenerator(fn func() string) ProviderOption {
	return func(p *SessionProvider) {
		p.generateSSID = fn
	}
}

// NewProvider 创建默认 SessionProvider。
//
// signingKey 会用于默认 JWTManager。若你通过 WithTokenManager 传入自己的实现，
// signingKey 可以为空，但为了配置清晰，生产环境仍建议显式传入。
func NewProvider(store Store, signingKey string, expiration time.Duration, opts ...ProviderOption) *SessionProvider {
	p := &SessionProvider{
		store:        store,
		manager:      NewJWTManager(expiration, signingKey),
		carrier:      defaultTokenCarrier{},
		expiration:   expiration,
		generateSSID: newSSID,
	}
	for _, opt := range opts {
		opt(p)
	}
	if p.generateSSID == nil {
		p.generateSSID = newSSID
	}
	return p
}

// NewSession 创建新的 Session，并把 token 写回响应。
func (p *SessionProvider) NewSession(ctx *gin.Context, uid int64, jwtData map[string]string, sessData map[string]any) (Session, error) {
	if err := p.checkReady(); err != nil {
		return nil, err
	}
	ssid := p.generateSSID()
	claims := Claims{
		Uid:        uid,
		SSID:       ssid,
		Data:       jwtData,
		Expiration: time.Now().Add(p.expiration).UnixMilli(),
	}
	token, err := p.manager.GenerateAccessToken(claims)
	if err != nil {
		return nil, err
	}
	p.carrier.Inject(ctx, token)
	if sessData == nil {
		sessData = map[string]any{}
	}
	sessData["uid"] = uid
	if err = p.store.Init(ctx.Request.Context(), ssid, sessData, p.expiration); err != nil {
		return nil, err
	}
	sess := newStoreSession(ssid, claims, p.store, p.expiration)
	ctx.Set(CtxSessionKey, sess)
	return sess, nil
}

// Get 从当前请求中获取并校验 Session。
func (p *SessionProvider) Get(ctx *gin.Context) (Session, error) {
	if err := p.checkReady(); err != nil {
		return nil, err
	}
	if val, ok := ctx.Get(CtxSessionKey); ok {
		if sess, ok := val.(Session); ok {
			return sess, nil
		}
	}
	token := p.carrier.Extract(ctx)
	if token == "" {
		return nil, ErrUnauthorized
	}
	claims, err := p.manager.VerifyAccessToken(token)
	if err != nil {
		return nil, err
	}
	if _, err = p.store.Get(ctx.Request.Context(), claims.SSID, "uid"); err != nil {
		return nil, fmt.Errorf("%w: 服务端 session 不存在或已过期: %v", ErrUnauthorized, err)
	}
	sess := newStoreSession(claims.SSID, claims, p.store, p.expiration)
	ctx.Set(CtxSessionKey, sess)
	return sess, nil
}

// Destroy 销毁当前请求对应的 Session，并清理响应侧 token。
func (p *SessionProvider) Destroy(ctx *gin.Context) error {
	sess, err := p.Get(ctx)
	if err != nil {
		return err
	}
	p.carrier.Clear(ctx)
	return sess.Destroy(ctx.Request.Context())
}

// UpdateClaims 重新签发当前 Session 的 JWT。
func (p *SessionProvider) UpdateClaims(ctx *gin.Context, claims Claims) error {
	if err := p.checkReady(); err != nil {
		return err
	}
	token, err := p.manager.GenerateAccessToken(claims)
	if err != nil {
		return err
	}
	p.carrier.Inject(ctx, token)
	return nil
}

// RenewAccessToken 使用当前请求的 token 重新签发一个新的 access token。
func (p *SessionProvider) RenewAccessToken(ctx *gin.Context) error {
	if err := p.checkReady(); err != nil {
		return err
	}
	token := p.carrier.Extract(ctx)
	if token == "" {
		return ErrUnauthorized
	}
	claims, err := p.manager.VerifyAccessToken(token)
	if err != nil {
		return err
	}
	claims.Expiration = time.Now().Add(p.expiration).UnixMilli()
	return p.UpdateClaims(ctx, claims)
}

func (p *SessionProvider) checkReady() error {
	if p == nil || p.store == nil || p.manager == nil || p.carrier == nil {
		return ErrDefaultProviderNotSet
	}
	return nil
}

type storeSession struct {
	ssid       string
	claims     Claims
	store      Store
	expiration time.Duration
}

func newStoreSession(ssid string, claims Claims, store Store, expiration time.Duration) *storeSession {
	return &storeSession{ssid: ssid, claims: claims, store: store, expiration: expiration}
}

// Set 将一个键值写入服务端 Session 存储。
func (s *storeSession) Set(ctx context.Context, key string, val any) error {
	return s.store.Set(ctx, s.ssid, key, val, s.expiration)
}

// Get 从服务端 Session 存储读取一个值。
func (s *storeSession) Get(ctx context.Context, key string) (any, error) {
	return s.store.Get(ctx, s.ssid, key)
}

// Del 删除服务端 Session 存储中的单个字段。
func (s *storeSession) Del(ctx context.Context, key string) error {
	return s.store.Del(ctx, s.ssid, key)
}

// Destroy 删除整个服务端 Session。
func (s *storeSession) Destroy(ctx context.Context) error {
	return s.store.Destroy(ctx, s.ssid)
}

// Claims 返回 JWT 中解析出来的身份声明。
func (s *storeSession) Claims() Claims {
	return s.claims
}

type defaultTokenCarrier struct{}

func (defaultTokenCarrier) Inject(ctx *gin.Context, value string) {
	ctx.Header("X-Access-Token", value)
}

func (defaultTokenCarrier) Extract(ctx *gin.Context) string {
	return ctx.GetHeader("Authorization")
}

func (defaultTokenCarrier) Clear(ctx *gin.Context) {
	ctx.Header("X-Access-Token", "")
}

func newSSID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b[:])
}
