package header

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// TestTokenCarrier 验证 Header carrier 可以注入和提取 Bearer token。
func TestTokenCarrier(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	ctx.Request.Header.Set("Authorization", "Bearer abc")

	carrier := NewTokenCarrier()
	if token := carrier.Extract(ctx); token != "abc" {
		t.Fatalf("unexpected token: %s", token)
	}
	carrier.Inject(ctx, "def")
	if got := recorder.Header().Get("X-Access-Token"); got != "def" {
		t.Fatalf("unexpected injected token: %s", got)
	}
}
