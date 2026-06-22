package ginx

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type testClaims struct {
	UID string
}

type testReq struct {
	Name string `json:"name" form:"name"`
}

// TestClaims 验证中间件写入的 claims 可以被 handler 按强类型取出.
func TestClaims(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/", nil)

	SetClaims(ctx, &testClaims{UID: "10001"})
	claims, ok := Claims[*testClaims](ctx)
	if !ok {
		t.Fatalf("expect claims")
	}
	if claims.UID != "10001" {
		t.Fatalf("unexpected uid: %s", claims.UID)
	}
}

// TestClaimsMissing 验证未写入 claims 时会返回 false.
func TestClaimsMissing(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/", nil)

	_, ok := Claims[*testClaims](ctx)
	if ok {
		t.Fatalf("expect missing claims")
	}
}

// TestWrapperClaims 验证 WrapperClaims 会把 gin.Context 中的 claims 注入业务处理函数.
func TestWrapperClaims(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	SetClaims(ctx, &testClaims{UID: "10001"})

	handler := WrapperClaims(func(ctx *gin.Context, claims *testClaims) (Result, error) {
		return Success.WithData(claims.UID), nil
	})
	handler(ctx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", recorder.Code)
	}
	var res Result
	if err := json.Unmarshal(recorder.Body.Bytes(), &res); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if res.Data != "10001" {
		t.Fatalf("unexpected data: %v", res.Data)
	}
}

// TestWrapperClaimsMissing 验证缺少 claims 时 WrapperClaims 会返回未授权.
func TestWrapperClaimsMissing(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/", nil)

	handler := WrapperClaims(func(ctx *gin.Context, claims *testClaims) (Result, error) {
		return Success, nil
	})
	handler(ctx)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("unexpected status: %d", recorder.Code)
	}
}

// TestWrapperClaimsAndBody 验证 WrapperClaimsAndBody 会同时注入 claims 和 body 参数.
func TestWrapperClaimsAndBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"name":"alice"}`))
	ctx.Request.Header.Set("Content-Type", "application/json")
	SetClaims(ctx, &testClaims{UID: "10001"})

	handler := WrapperClaimsAndBody(func(ctx *gin.Context, claims *testClaims, req testReq) (Result, error) {
		return Success.WithData(claims.UID + ":" + req.Name), nil
	})
	handler(ctx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", recorder.Code)
	}
	var res Result
	if err := json.Unmarshal(recorder.Body.Bytes(), &res); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if res.Data != "10001:alice" {
		t.Fatalf("unexpected data: %v", res.Data)
	}
}

// TestWrapperClaimsAndQuery 验证 WrapperClaimsAndQuery 会同时注入 claims 和 query 参数.
func TestWrapperClaimsAndQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/?name=bob", nil)
	SetClaims(ctx, &testClaims{UID: "10001"})

	handler := WrapperClaimsAndQuery(func(ctx *gin.Context, claims *testClaims, req testReq) (Result, error) {
		return Success.WithData(claims.UID + ":" + req.Name), nil
	})
	handler(ctx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", recorder.Code)
	}
	var res Result
	if err := json.Unmarshal(recorder.Body.Bytes(), &res); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if res.Data != "10001:bob" {
		t.Fatalf("unexpected data: %v", res.Data)
	}
}
