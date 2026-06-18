package session

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	// CtxSessionKey 是 Session 写入 gin.Context 的默认 key。
	CtxSessionKey = "_session"
)

var (
	defaultProviderMu sync.RWMutex
	defaultProvider   Provider
)

// SetDefaultProvider 设置全局默认 Provider。
//
// 全局 Provider 主要服务于 NewSession、Get、CheckLoginMiddleware 这类快捷方法。
// 如果一个项目里存在多套登录体系，更推荐显式创建 MiddlewareBuilder 并传入对应 Provider。
func SetDefaultProvider(provider Provider) {
	defaultProviderMu.Lock()
	defer defaultProviderMu.Unlock()
	defaultProvider = provider
}

// DefaultProvider 返回当前全局默认 Provider。
func DefaultProvider() Provider {
	defaultProviderMu.RLock()
	defer defaultProviderMu.RUnlock()
	return defaultProvider
}

// NewSession 使用全局默认 Provider 创建 Session。
func NewSession(ctx *gin.Context, uid int64, jwtData map[string]string, sessData map[string]any) (Session, error) {
	provider := DefaultProvider()
	if provider == nil {
		return nil, ErrDefaultProviderNotSet
	}
	return provider.NewSession(ctx, uid, jwtData, sessData)
}

// Get 使用全局默认 Provider 从当前请求获取 Session。
func Get(ctx *gin.Context) (Session, error) {
	if val, ok := ctx.Get(CtxSessionKey); ok {
		if sess, ok := val.(Session); ok {
			return sess, nil
		}
	}
	provider := DefaultProvider()
	if provider == nil {
		return nil, ErrDefaultProviderNotSet
	}
	return provider.Get(ctx)
}

// Destroy 使用全局默认 Provider 销毁当前请求的 Session。
func Destroy(ctx *gin.Context) error {
	provider := DefaultProvider()
	if provider == nil {
		return ErrDefaultProviderNotSet
	}
	return provider.Destroy(ctx)
}

// RenewAccessToken 使用全局默认 Provider 刷新 access token。
func RenewAccessToken(ctx *gin.Context) error {
	provider := DefaultProvider()
	if provider == nil {
		return ErrDefaultProviderNotSet
	}
	return provider.RenewAccessToken(ctx)
}

// UpdateClaims 使用全局默认 Provider 更新 JWT Claims。
func UpdateClaims(ctx *gin.Context, claims Claims) error {
	provider := DefaultProvider()
	if provider == nil {
		return ErrDefaultProviderNotSet
	}
	return provider.UpdateClaims(ctx, claims)
}

// CheckLoginMiddleware 创建使用全局默认 Provider 的登录校验中间件。
func CheckLoginMiddleware() gin.HandlerFunc {
	return (&MiddlewareBuilder{
		Provider:  DefaultProvider(),
		Threshold: 30 * time.Minute,
	}).Build()
}
