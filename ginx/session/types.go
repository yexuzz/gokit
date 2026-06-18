package session

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
)

// Session 表示一次已经通过校验的服务端会话。
//
// 这个接口混合了传统 Session 和 JWT 的设计：
//  1. Claims 保存适合放进 JWT 的少量身份信息，例如 uid、ssid、过期时间和少量业务标记。
//  2. Set/Get/Del 操作的是服务端存储里的 Session 数据，例如临时状态、权限快照或登录设备信息。
//
// 注意：Get 只读取服务端 Session 数据，不会读取 JWT Claims 里的 Data。
// 如果要读取 JWT 中的字段，应该通过 sess.Claims().Get("key")。
type Session interface {
	// Set 将一个键值写入服务端 Session 存储。
	//
	// 对 Redis 这类远端存储来说，这里通常会写入 hash；对内存存储来说，会写入 map。
	// val 使用 any，避免绑定 ekit.AnyValue 或某个具体序列化实现。
	Set(ctx context.Context, key string, val any) error

	// Get 从服务端 Session 存储读取一个值。
	//
	// 当 key 不存在、Session 已过期或底层存储不可用时，返回 error。
	// 返回值类型取决于底层存储：内存存储会保留原始类型，Redis 存储通常返回 string。
	Get(ctx context.Context, key string) (any, error)

	// Del 删除服务端 Session 存储中的单个字段。
	Del(ctx context.Context, key string) error

	// Destroy 销毁整个 Session，通常用于退出登录。
	Destroy(ctx context.Context) error

	// Claims 返回从 JWT 中解析出来的身份声明。
	Claims() Claims
}

// Provider 定义 Session 的完整生命周期管理能力。
//
// Provider 负责四件事：
//  1. 登录成功后创建 Session，并把 access token 写回 cookie/header。
//  2. 请求进入系统时从 cookie/header 中提取 token，并校验 JWT。
//  3. 在需要时刷新 JWT claims，因为 JWT 本身不可变，刷新本质上是重新签发 token。
//  4. 退出登录时清理 token 和服务端 Session 数据。
type Provider interface {
	// NewSession 创建新的 Session。
	//
	// jwtData 会被编码到 JWT Claims.Data 中，适合存少量不敏感、需要频繁读取的身份信息。
	// sessData 会被写入服务端存储，适合存体积更大或需要服务端主动删除的数据。
	NewSession(ctx *gin.Context, uid int64, jwtData map[string]string, sessData map[string]any) (Session, error)

	// Get 从当前 gin.Context 中获取并校验 Session。
	//
	// 实现必须保证返回的 Session 已经通过 token 校验，并且服务端 Session 仍然存在。
	Get(ctx *gin.Context) (Session, error)

	// Destroy 销毁当前请求对应的 Session。
	Destroy(ctx *gin.Context) error

	// UpdateClaims 更新 JWT Claims。
	//
	// 因为 JWT 签发后不可修改，所以这里会重新生成 token 并写回 TokenCarrier。
	UpdateClaims(ctx *gin.Context, claims Claims) error

	// RenewAccessToken 刷新当前请求携带的 access token。
	RenewAccessToken(ctx *gin.Context) error
}

// Claims 是写入 JWT 的会话声明。
//
// 它只应该保存体积小、变化不频繁、即使客户端可见也不会造成安全问题的数据。
// 需要服务端强控制的数据应该放到 Session 存储里，而不是放到 JWT 里。
type Claims struct {
	// Uid 表示登录用户 ID。
	Uid int64 `json:"uid"`

	// SSID 表示服务端 Session ID。
	//
	// JWT 用它定位 Redis 或内存中的服务端 Session 数据。
	SSID string `json:"ssid"`

	// Data 表示额外写入 JWT 的少量业务数据。
	Data map[string]string `json:"data,omitempty"`

	// Expiration 表示 JWT 和服务端 Session 的过期时间，单位是毫秒时间戳。
	Expiration int64 `json:"expiration"`
}

// Get 从 Claims.Data 中读取一个值。
//
// 返回 any 是为了和 Session.Get 的使用方式保持一致，但实际值来自 map[string]string，
// 因此成功时底层类型一定是 string。
func (c Claims) Get(key string) (any, error) {
	if c.Data == nil {
		return nil, ErrSessionKeyNotFound
	}
	val, ok := c.Data[key]
	if !ok {
		return nil, ErrSessionKeyNotFound
	}
	return val, nil
}

// TokenCarrier 负责 token 的输入输出位置。
//
// 例如：
//  1. header.TokenCarrier 从 Authorization 中读取 token，并把新 token 写入响应头。
//  2. cookie.TokenCarrier 从 Cookie 中读取 token，并把新 token 写入 Set-Cookie。
//  3. mixin.TokenCarrier 可以组合多个 carrier，同时支持 header 和 cookie。
type TokenCarrier interface {
	// Inject 将 token 写入响应。
	Inject(ctx *gin.Context, value string)

	// Extract 从请求中读取 token。
	Extract(ctx *gin.Context) string

	// Clear 清理响应侧 token。
	Clear(ctx *gin.Context)
}

// Store 定义服务端 Session 数据的存储能力。
//
// Store 不关心 JWT，也不关心 token 放在 header 还是 cookie。
// 它只负责根据 ssid 读写服务端 Session 数据，因此可以有内存、Redis、数据库等不同实现。
type Store interface {
	// Init 初始化一个新的 Session 数据集合，并设置过期时间。
	Init(ctx context.Context, ssid string, data map[string]any, expiration time.Duration) error

	// Set 写入 Session 中的单个字段。
	Set(ctx context.Context, ssid string, key string, val any, expiration time.Duration) error

	// Get 读取 Session 中的单个字段。
	Get(ctx context.Context, ssid string, key string) (any, error)

	// Del 删除 Session 中的单个字段。
	Del(ctx context.Context, ssid string, key string) error

	// Destroy 删除整个 Session。
	Destroy(ctx context.Context, ssid string) error
}
