package ginx

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// WrapperNone 包装不需要自动绑定请求参数的 gin 处理函数。
func WrapperNone(fn func(ctx *gin.Context) (Result, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		result, err := fn(ctx)
		if err != nil {
			handleWrapperError(ctx, err)
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}

// WrapperBody 包装需要从 body 绑定请求参数的 gin 处理函数。
func WrapperBody[Req any](fn func(ctx *gin.Context, req Req) (Result, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req Req
		if err := ctx.ShouldBindJSON(&req); err != nil {
			// 请求体绑定失败属于客户端参数错误，直接返回 400
			logWarn(ctx, "序列化参数失败", "err", err, "method", ctx.Request.Method, "path", ctx.FullPath())
			ctx.JSON(http.StatusBadRequest, NewResult(http.StatusBadRequest, "bad request"))
			return
		}
		result, err := fn(ctx, req)
		if err != nil {
			handleWrapperError(ctx, err)
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}

// WrapperQuery 包装需要从 query 绑定请求参数的 gin 处理函数。
func WrapperQuery[Req any](fn func(ctx *gin.Context, req Req) (Result, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req Req
		if err := ctx.ShouldBindQuery(&req); err != nil {
			// 请求体绑定失败属于客户端参数错误，直接返回 400
			logWarn(ctx, "序列化参数失败", "err", err, "method", ctx.Request.Method, "path", ctx.FullPath())
			ctx.JSON(http.StatusBadRequest, NewResult(http.StatusBadRequest, "bad request"))
			return
		}
		result, err := fn(ctx, req)
		if err != nil {
			handleWrapperError(ctx, err)
			return
		}
		ctx.JSON(http.StatusOK, result)
	}
}

// handleWrapperError 统一处理 ginx 包装器中的业务错误和内部错误。
func handleWrapperError(ctx *gin.Context, err error) {
	if codedErr, ok := AsCodedError(err); ok {
		// 业务错误可以安全返回给前端，不记录为系统内部错误。
		ctx.JSON(http.StatusOK, NewResult(codedErr.ErrorCode(), codedErr.ErrorMessage()))
		return
	}
	// 未声明为业务错误的异常统一视为内部错误，并交给注入的日志出口记录。
	logError(ctx, "系统内部发生错误", "err", err, "method", ctx.Request.Method, "path", ctx.FullPath())
	ctx.JSON(http.StatusInternalServerError, Fail)
}
