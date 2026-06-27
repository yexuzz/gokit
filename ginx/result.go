package ginx

import "net/http"

var (
	Success = Result{Code: http.StatusOK, Message: "success"}
	Fail    = Result{Code: http.StatusInternalServerError, Message: "fail"}
)

type Result struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
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
