package session

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// MiddlewareBuilder 用于构造登录校验中间件。
type MiddlewareBuilder struct {
	// Provider 是当前中间件使用的 Session Provider。
	Provider Provider

	// Threshold 表示 token 剩余有效期小于该值时，自动刷新 access token。
	//
	// 如果为 0，则不会自动刷新 token。
	Threshold time.Duration

	// Unauthorized 允许调用方自定义未授权响应。
	//
	// 如果为 nil，默认返回 HTTP 401。
	Unauthorized func(ctx *gin.Context, err error)
}

// Build 创建 gin 登录校验中间件。
func (b *MiddlewareBuilder) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		provider := b.Provider
		if provider == nil {
			provider = DefaultProvider()
		}
		if provider == nil {
			b.abort(ctx, ErrDefaultProviderNotSet)
			return
		}
		sess, err := provider.Get(ctx)
		if err != nil {
			b.abort(ctx, err)
			return
		}
		if b.Threshold > 0 {
			now := time.Now().UnixMilli()
			if sess.Claims().Expiration-now < b.Threshold.Milliseconds() {
				_ = provider.RenewAccessToken(ctx)
			}
		}
		ctx.Set(CtxSessionKey, sess)
	}
}

func (b *MiddlewareBuilder) abort(ctx *gin.Context, err error) {
	if b.Unauthorized != nil {
		b.Unauthorized(ctx, err)
		return
	}
	ctx.AbortWithStatus(http.StatusUnauthorized)
}
