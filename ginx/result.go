package ginx

import "net/http"

const (
	CodeOK              = http.StatusOK                  // CodeOK 表示请求处理成功.
	CodeInvalidParam    = http.StatusBadRequest          // CodeInvalidParam 表示请求参数不合法.
	CodeUnauthorized    = http.StatusUnauthorized        // CodeUnauthorized 表示用户未登录或登录 token 无效.
	CodeForbidden       = http.StatusForbidden           // CodeForbidden 表示用户没有权限访问资源或执行操作.
	CodeTooManyRequests = http.StatusTooManyRequests     // CodeTooManyRequests 表示请求过于频繁.
	CodeInternal        = http.StatusInternalServerError // CodeInternal 表示系统内部错误.
)

var (
	Success = Result{Code: CodeOK, Message: "success"}
	Fail    = Result{Code: CodeInternal, Message: "fail"}
)

type Result struct {
	Code    int    `json:"code"`           // 业务状态码 200-成功,401-未登录,403-没有权限,429-请求过于频繁,500-服务内部错误
	Message string `json:"message"`        // 错误信息
	Data    any    `json:"data,omitempty"` // 返回数据
}

func NewResult(code int, message string) Result {
	return Result{Code: code, Message: message}
}

func (r Result) WithData(data any) Result {
	r.Data = data
	return r
}

type PageRequest struct {
	Page     int `form:"page" json:"page"`
	PageSize int `form:"page_size" json:"page_size"`
}

type PageResponse[T any] struct {
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
	Total    int64 `json:"total"`
	Items    []T   `json:"items"`
}
