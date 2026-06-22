package jwtx

import "errors"

var (
	// ErrInvalidToken 表示 token 为空, 验签失败, 过期或 claims 类型不正确.
	ErrInvalidToken = errors.New("jwtx: invalid token")
	// ErrSessionRevoked 表示 token 所属的登录会话已经被主动失效.
	ErrSessionRevoked = errors.New("jwtx: session revoked")
	// ErrRefreshTokenInvalid 表示 refresh token 没有通过服务端状态校验.
	ErrRefreshTokenInvalid = errors.New("jwtx: refresh token invalid")
	// ErrClaimsFactoryRequired 表示创建 Handler 时没有提供 claims 工厂函数.
	ErrClaimsFactoryRequired = errors.New("jwtx: claims factory required")
)
