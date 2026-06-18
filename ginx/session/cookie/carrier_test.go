package cookie

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// TestTokenCarrier 验证 Cookie carrier 可以注入和提取 token。
func TestTokenCarrier(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: "abc"})
	ctx.Request = req

	carrier := NewTokenCarrier("access_token", 3600)
	if token := carrier.Extract(ctx); token != "abc" {
		t.Fatalf("unexpected token: %s", token)
	}
	carrier.Inject(ctx, "def")
	if len(recorder.Result().Cookies()) == 0 {
		t.Fatalf("expect cookie")
	}
}
