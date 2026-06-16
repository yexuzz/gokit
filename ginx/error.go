package ginx

import "errors"

// CodedError 表示可以转换成接口响应码和响应文案的错误。
type CodedError interface {
	error
	ErrorCode() int
	ErrorMessage() string
}

func asCodedError(err error) (CodedError, bool) {
	if err == nil {
		return nil, false
	}
	var codedErr CodedError
	if errors.As(err, &codedErr) {
		return codedErr, true
	}
	return nil, false
}

// AsCodedError 将业务错误转换为可返回给前端的错误码和文案。
func AsCodedError(err error) (CodedError, bool) {
	return asCodedError(err)
}
