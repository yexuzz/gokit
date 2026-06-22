package ginx

import "github.com/gin-gonic/gin"

const (
	// CtxClaimsKey 表示 gin.Context 中保存登录用户 claims 的默认 key.
	CtxClaimsKey = "user"
)

// SetClaims 将登录中间件校验出来的 claims 写入 gin.Context.
//
// 这个方法只负责保存当前请求内的用户信息, 不关心 claims 来自 JWT, Session 还是其它认证方式.
func SetClaims(ctx *gin.Context, claims any) {
	ctx.Set(CtxClaimsKey, claims)
}

// Claims 从 gin.Context 中读取登录用户 claims.
//
// JWT 中间件可以先调用 SetClaims 写入业务自定义 claims, 后续 handler 再通过泛型取出强类型数据.
func Claims[T any](ctx *gin.Context) (T, bool) {
	val, ok := ctx.Get(CtxClaimsKey)
	if !ok {
		var zero T
		return zero, false
	}
	claims, ok := val.(T)
	return claims, ok
}
