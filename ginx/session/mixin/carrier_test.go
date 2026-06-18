package mixin

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/yexuzz/gokit/ginx/session/header"
)

// TestTokenCarrier 验证组合 carrier 会按顺序提取 token。
func TestTokenCarrier(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	ctx.Request.Header.Set("Authorization", "Bearer abc")

	carrier := NewTokenCarrier(header.NewTokenCarrier())
	if token := carrier.Extract(ctx); token != "abc" {
		t.Fatalf("unexpected token: %s", token)
	}
}
