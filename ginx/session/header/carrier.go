package header

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/yexuzz/gokit/ginx/session"
)

var _ session.TokenCarrier = (*TokenCarrier)(nil)

// TokenCarrier 使用 HTTP Header 携带 access token。
//
// 读取时默认从 Authorization 中解析 Bearer token。
// 写入时默认把新 token 暴露到 X-Access-Token，方便前端在响应后更新本地 token。
type TokenCarrier struct {
	// InputName 是请求头名称，默认 Authorization。
	InputName string

	// OutputName 是响应头名称，默认 X-Access-Token。
	OutputName string

	// Prefix 是请求头中的 token 前缀，默认 Bearer。
	Prefix string
}

// NewTokenCarrier 创建默认 Header token carrier。
func NewTokenCarrier() *TokenCarrier {
	return &TokenCarrier{
		InputName:  "Authorization",
		OutputName: "X-Access-Token",
		Prefix:     "Bearer",
	}
}

// Inject 将 token 写入响应头。
func (t *TokenCarrier) Inject(ctx *gin.Context, value string) {
	ctx.Header(t.outputName(), value)
}

// Extract 从请求头中读取 token。
func (t *TokenCarrier) Extract(ctx *gin.Context) string {
	raw := ctx.GetHeader(t.inputName())
	if raw == "" {
		return ""
	}
	prefix := t.prefix()
	if prefix == "" {
		return strings.TrimSpace(raw)
	}
	return strings.TrimSpace(strings.TrimPrefix(raw, prefix+" "))
}

// Clear 清空响应头中的 token。
func (t *TokenCarrier) Clear(ctx *gin.Context) {
	ctx.Header(t.outputName(), "")
}

func (t *TokenCarrier) inputName() string {
	if t.InputName == "" {
		return "Authorization"
	}
	return t.InputName
}

func (t *TokenCarrier) outputName() string {
	if t.OutputName == "" {
		return "X-Access-Token"
	}
	return t.OutputName
}

func (t *TokenCarrier) prefix() string {
	if t.Prefix == "" {
		return "Bearer"
	}
	return t.Prefix
}
