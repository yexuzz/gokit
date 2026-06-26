package jwtx

import "errors"

var (
	// ErrInvalidToken 表示 token 为空, 验签失败, 过期或声明结构不正确.
	ErrInvalidToken = errors.New("jwtx: invalid token")
	// ErrSessionRevoked 表示 token 所属的登录会话已经被主动失效.
	ErrSessionRevoked = errors.New("jwtx: session revoked")
	// ErrRefreshTokenInvalid 表示 refresh token 没有通过服务端状态校验.
	ErrRefreshTokenInvalid = errors.New("jwtx: refresh token invalid")
	// ErrUserSessionStoreUnsupported 表示当前 Store 不支持按用户管理多设备会话集合.
	ErrUserSessionStoreUnsupported = errors.New("jwtx: user session store unsupported")
)
