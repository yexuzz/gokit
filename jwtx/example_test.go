package jwtx_test

import (
	"context"
	"time"

	"github.com/yexuzz/gokit/jwtx"
)

// adminClaims 模拟业务项目自己的后台用户 claims.
type adminClaims struct {
	jwtx.BaseClaims
	UID      string `json:"uid"`
	Username string `json:"username"`
}

// ExampleDefaultHandler 展示业务项目如何把 jwtx 当成不依赖 Gin 的 JWT Handler 使用.
func ExampleDefaultHandler() {
	handler := jwtx.MustNewHandler(func() *adminClaims { return &adminClaims{} },
		jwtx.WithAccessTokenKey[*adminClaims]([]byte("access-secret")),
		jwtx.WithRefreshTokenKey[*adminClaims]([]byte("refresh-secret")),
		jwtx.WithExpiration[*adminClaims](15*time.Minute, 7*24*time.Hour),
	)

	tokens, _ := handler.SetLoginToken(context.Background(), &adminClaims{
		UID:      "10001",
		Username: "admin",
	})
	_, _ = handler.CheckAccessToken(context.Background(), tokens.AccessToken)
}
