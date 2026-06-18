package mixin

import (
	"github.com/gin-gonic/gin"
	"github.com/yexuzz/gokit/ginx/session"
)

var _ session.TokenCarrier = (*TokenCarrier)(nil)

// TokenCarrier 组合多个 TokenCarrier。
//
// 常见用法是同时支持 header 和 cookie：
//  1. Inject 会把新 token 写入所有 carrier。
//  2. Extract 会按顺序读取，谁先读到有效 token 就使用谁。
//  3. Clear 会清理所有 carrier。
type TokenCarrier struct {
	carriers []session.TokenCarrier
}

// NewTokenCarrier 创建组合 carrier。
func NewTokenCarrier(carriers ...session.TokenCarrier) *TokenCarrier {
	return &TokenCarrier{carriers: carriers}
}

// Inject 将 token 写入所有 carrier。
func (t *TokenCarrier) Inject(ctx *gin.Context, value string) {
	for _, carrier := range t.carriers {
		carrier.Inject(ctx, value)
	}
}

// Extract 按顺序从 carrier 中读取 token。
func (t *TokenCarrier) Extract(ctx *gin.Context) string {
	for _, carrier := range t.carriers {
		val := carrier.Extract(ctx)
		if val != "" {
			return val
		}
	}
	return ""
}

// Clear 清理所有 carrier。
func (t *TokenCarrier) Clear(ctx *gin.Context) {
	for _, carrier := range t.carriers {
		carrier.Clear(ctx)
	}
}
