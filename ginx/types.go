package ginx

import (
	"github.com/gin-gonic/gin"
	jwtv5 "github.com/golang-jwt/jwt/v5"
)

type RegisterRoutes interface {
	RegisterPrivate(server *gin.Engine)
	RegisterPublic(server *gin.Engine)
}

type UserClaims struct {
	UserID    string `json:"user_id"`
	UserAgent string `json:"user_agent"`
	Ssid      string `json:"ssid"`
	jwtv5.RegisteredClaims
}
