package session

import "github.com/gin-gonic/gin"

// Builder 用于分步骤创建 Session。
//
// 它适合登录接口中逐步追加 jwtData 和 sessData，最后统一调用 Build。
// 如果数据很简单，也可以直接调用 provider.NewSession。
type Builder struct {
	ctx      *gin.Context
	uid      int64
	jwtData  map[string]string
	sessData map[string]any
	provider Provider
}

// NewSessionBuilder 创建 Session Builder。
//
// 默认使用全局 Provider；如果当前登录流程要使用特殊 Provider，可以调用 SetProvider 覆盖。
func NewSessionBuilder(ctx *gin.Context, uid int64) *Builder {
	return &Builder{
		ctx:      ctx,
		uid:      uid,
		provider: DefaultProvider(),
	}
}

// SetProvider 设置当前 Builder 使用的 Provider。
func (b *Builder) SetProvider(provider Provider) *Builder {
	b.provider = provider
	return b
}

// SetJWTData 设置要写入 JWT Claims.Data 的数据。
func (b *Builder) SetJWTData(data map[string]string) *Builder {
	b.jwtData = data
	return b
}

// SetSessData 设置要写入服务端 Session 存储的数据。
func (b *Builder) SetSessData(data map[string]any) *Builder {
	b.sessData = data
	return b
}

// Build 创建 Session。
func (b *Builder) Build() (Session, error) {
	if b.provider == nil {
		return nil, ErrDefaultProviderNotSet
	}
	return b.provider.NewSession(b.ctx, b.uid, b.jwtData, b.sessData)
}
