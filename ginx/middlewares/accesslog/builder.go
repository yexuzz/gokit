package accesslog

import (
	"bytes"
	"context"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/atomic"
)

type AccessLog struct {
	Method   string // http 请求类型
	Url      string // url 整个请求的url
	ReqBody  string // 请求体
	RespBody string // 响应体
	Duration string // 处理时间
	Status   int    // 状态码
}

type Builder struct {
	allowReqBody  *atomic.Bool
	allowRespBody *atomic.Bool
	loggerFunc    func(ctx context.Context, al *AccessLog)
	maxLength     *atomic.Int64
}

func NewBuilder(fn func(ctx context.Context, al *AccessLog)) *Builder {
	return &Builder{
		allowReqBody:  atomic.NewBool(false),
		allowRespBody: atomic.NewBool(false),
		loggerFunc:    fn,
		maxLength:     atomic.NewInt64(1024),
	}
}

// AllowReqBody 是否打印请求体
func (b *Builder) AllowReqBody() *Builder {
	b.allowReqBody.Store(true)
	return b
}

// AllowRespBody 是否打印响应体
func (b *Builder) AllowRespBody() *Builder {
	b.allowRespBody.Store(true)
	return b
}

// MaxLength 打印的最大长度
func (b *Builder) MaxLength(maxLength int64) *Builder {
	b.maxLength.Store(maxLength)
	return b
}

func (b *Builder) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var (
			//请求处理开始时间
			start = time.Now()
			//url
			url = ctx.Request.URL.String()
			//url 长度
			curLen = int64(len(url))
			//运行打印的最大长度
			maxLength = b.maxLength.Load()
			//是否打印请求体
			allowReqBody = b.allowReqBody.Load()
			//是否打印响应体
			allowRespBody = b.allowRespBody.Load()
		)

		if curLen >= maxLength {
			url = url[:maxLength]
		}

		accessLog := &AccessLog{
			Method: ctx.Request.Method,
			Url:    url,
		}
		if ctx.Request.Body != nil && allowReqBody {
			body, _ := ctx.GetRawData()
			ctx.Request.Body = io.NopCloser(bytes.NewReader(body))
			if int64(len(body)) >= maxLength {
				body = body[:maxLength]
			}
			//注意资源的消耗
			accessLog.ReqBody = string(body)
		}

		if allowRespBody {
			ctx.Writer = responseWriter{
				ResponseWriter: ctx.Writer,
				al:             accessLog,
				maxLength:      maxLength,
			}
		}

		defer func() {
			accessLog.Duration = time.Since(start).String()
			//日志打印
			b.loggerFunc(ctx, accessLog)
		}()
		ctx.Next()
	}
}

type responseWriter struct {
	gin.ResponseWriter
	al        *AccessLog
	maxLength int64
}

func (r responseWriter) WriteHeader(statusCode int) {
	r.al.Status = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func (r responseWriter) Write(data []byte) (int, error) {
	curLen := int64(len(data))
	if curLen >= r.maxLength {
		data = data[:r.maxLength]
	}
	r.al.RespBody = string(data)
	return r.ResponseWriter.Write(data)
}
