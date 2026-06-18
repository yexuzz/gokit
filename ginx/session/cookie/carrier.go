package cookie

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yexuzz/gokit/ginx/session"
)

var _ session.TokenCarrier = (*TokenCarrier)(nil)

// TokenCarrier 使用 Cookie 携带 access token。
//
// Cookie 方案适合浏览器应用，尤其是配合 HttpOnly 减少前端脚本直接读取 token 的风险。
// 如果前后端跨域，需要同时正确配置 SameSite、Secure、Domain 和 CORS。
type TokenCarrier struct {
	// MaxAge 表示 Cookie 最大存活时间，单位秒。
	MaxAge int

	// Name 表示 Cookie 名称。
	Name string

	// Path 表示 Cookie 生效路径。
	Path string

	// Domain 表示 Cookie 生效域名。
	Domain string

	// Secure 表示是否只允许 HTTPS 传输。
	Secure bool

	// HttpOnly 表示是否禁止浏览器脚本读取 Cookie。
	HttpOnly bool

	// SameSite 表示 Cookie SameSite 策略。
	SameSite http.SameSite
}

// NewTokenCarrier 创建默认 Cookie token carrier。
func NewTokenCarrier(name string, maxAge int) *TokenCarrier {
	if name == "" {
		name = "access_token"
	}
	return &TokenCarrier{
		MaxAge:   maxAge,
		Name:     name,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
}

// Inject 将 token 写入 Cookie。
func (t *TokenCarrier) Inject(ctx *gin.Context, value string) {
	http.SetCookie(ctx.Writer, &http.Cookie{
		Name:     t.name(),
		Value:    value,
		Path:     t.path(),
		Domain:   t.Domain,
		MaxAge:   t.MaxAge,
		Secure:   t.Secure,
		HttpOnly: t.HttpOnly,
		SameSite: t.sameSite(),
	})
}

// Extract 从 Cookie 中读取 token。
func (t *TokenCarrier) Extract(ctx *gin.Context) string {
	val, err := ctx.Cookie(t.name())
	if err != nil {
		return ""
	}
	return val
}

// Clear 通过设置 MaxAge=-1 清理 Cookie。
func (t *TokenCarrier) Clear(ctx *gin.Context) {
	http.SetCookie(ctx.Writer, &http.Cookie{
		Name:     t.name(),
		Value:    "",
		Path:     t.path(),
		Domain:   t.Domain,
		MaxAge:   -1,
		Secure:   t.Secure,
		HttpOnly: t.HttpOnly,
		SameSite: t.sameSite(),
	})
}

func (t *TokenCarrier) name() string {
	if t.Name == "" {
		return "access_token"
	}
	return t.Name
}

func (t *TokenCarrier) path() string {
	if t.Path == "" {
		return "/"
	}
	return t.Path
}

func (t *TokenCarrier) sameSite() http.SameSite {
	if t.SameSite == 0 {
		return http.SameSiteLaxMode
	}
	return t.SameSite
}
