package jwt

import (
	"github.com/gin-gonic/gin"
	"github.com/yexuzz/gokit/ginx"
	"github.com/yexuzz/gokit/jwtx"
)

// LoginMiddlewareBuilder 构建 Gin 登录态校验中间件.
//
// 它只负责中间件编排: 忽略不需要登录的路径, 调用 Handler 校验 token,
// 并把校验后的 session 写入 gin.Context.
type LoginMiddlewareBuilder[T any] struct {
	handler     *Handler[T]
	ignorePaths map[string]struct{}
}

// NewLoginMiddlewareBuilder 创建 Gin JWT 登录态中间件构建器.
func NewLoginMiddlewareBuilder[T any](handler *Handler[T]) *LoginMiddlewareBuilder[T] {
	return &LoginMiddlewareBuilder[T]{
		handler:     handler,
		ignorePaths: make(map[string]struct{}),
	}
}

// IgnorePaths 追加不需要登录态校验的路由路径.
func (b *LoginMiddlewareBuilder[T]) IgnorePaths(paths ...string) *LoginMiddlewareBuilder[T] {
	for _, path := range paths {
		b.ignorePaths[path] = struct{}{}
	}
	return b
}

// Build 创建 Gin 登录态校验中间件.
//
// 校验通过后会把 jwtx.Session 写入 gin.Context, 后续可以通过 ginx.WrapperClaims 读取.
func (b *LoginMiddlewareBuilder[T]) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if _, ok := b.ignorePaths[ctx.Request.URL.Path]; ok {
			ctx.Next()
			return
		}

		session, err := b.handler.CheckToken(ctx)
		if err != nil {
			b.handler.cfg.ErrorHandler(ctx, err)
			return
		}
		ginx.SetClaims(ctx, session)
		ctx.Next()
	}
}

// Session 从 gin.Context 中读取当前登录会话.
func Session[T any](ctx *gin.Context) (jwtx.Session[T], bool) {
	return ginx.Claims[jwtx.Session[T]](ctx)
}

// Payload 从 gin.Context 中读取当前登录用户的业务载荷.
func Payload[T any](ctx *gin.Context) (T, bool) {
	session, ok := Session[T](ctx)
	if !ok {
		var zero T
		return zero, false
	}
	return session.Payload, true
}
