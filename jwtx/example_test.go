package jwtx_test

import (
	"context"
	"time"

	"github.com/yexuzz/gokit/jwtx"
)

// adminPayload 模拟业务项目自己的后台用户登录数据.
type adminPayload struct {
	UID      string `json:"uid"`
	Username string `json:"username"`
}

// ExampleDefaultManager 展示业务项目如何把 jwtx 当成不依赖 Gin 的 JWT Manager 使用.
func ExampleDefaultManager() {
	manager := jwtx.MustNewManager[adminPayload](
		jwtx.WithAccessTokenKey([]byte("access-secret")),
		jwtx.WithRefreshTokenKey([]byte("refresh-secret")),
		jwtx.WithExpiration(15*time.Minute, 7*24*time.Hour),
	)

	tokens, _ := manager.SetLoginToken(context.Background(), adminPayload{
		UID:      "10001",
		Username: "admin",
	})
	session, _ := manager.CheckAccessToken(context.Background(), tokens.AccessToken)
	_ = session.Payload.UID
}
