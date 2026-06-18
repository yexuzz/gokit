package session

import "errors"

var (
	// ErrUnauthorized 表示当前请求没有通过 Session 校验。
	//
	// 常见原因包括：没有携带 token、token 签名错误、token 过期、服务端 Session 已被删除。
	ErrUnauthorized = errors.New("未授权")

	// ErrSessionKeyNotFound 表示 Session 或 Claims 中没有找到指定 key。
	ErrSessionKeyNotFound = errors.New("session 中没有找到对应的 key")

	// ErrDefaultProviderNotSet 表示使用全局快捷方法前没有设置默认 Provider。
	ErrDefaultProviderNotSet = errors.New("session 默认 provider 未设置")
)
