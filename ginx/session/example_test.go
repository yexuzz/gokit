package session_test

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yexuzz/gokit/ginx/session"
	"github.com/yexuzz/gokit/ginx/session/header"
)

// ExampleNewProvider 展示生产代码里最常见的初始化方式。
func ExampleNewProvider() {
	store := session.NewMemoryStore()
	provider := session.NewProvider(
		store,
		"change-me-to-a-strong-secret",
		2*time.Hour,
		session.WithTokenCarrier(header.NewTokenCarrier()),
	)
	session.SetDefaultProvider(provider)

	router := gin.New()
	router.POST("/login", func(ctx *gin.Context) {
		_, _ = session.NewSession(ctx, 10001,
			map[string]string{"role": "admin"},
			map[string]any{"device": "terminal"},
		)
		ctx.Status(http.StatusNoContent)
	})
	router.GET("/profile", session.CheckLoginMiddleware(), func(ctx *gin.Context) {
		sess, _ := session.Get(ctx)
		ctx.JSON(http.StatusOK, gin.H{"uid": sess.Claims().Uid})
	})
}
