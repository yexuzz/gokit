package logx

import "context"

// ContextKey 表示 logx 写入上下文的字段名。
type ContextKey string

const (
	// TraceIDKey 表示链路追踪 ID 在上下文和日志中的字段名。
	TraceIDKey ContextKey = "trace_id"
	// RequestIDKey 表示请求 ID 在上下文和日志中的字段名。
	RequestIDKey ContextKey = "request_id"
)

// WithTraceID 将链路追踪 ID 写入上下文。
func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, TraceIDKey, traceID)
}

// WithRequestID 将请求 ID 写入上下文。
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, RequestIDKey, requestID)
}

// TraceID 从上下文读取链路追踪 ID。
func TraceID(ctx context.Context) string {
	return contextString(ctx, TraceIDKey)
}

// RequestID 从上下文读取请求 ID。
func RequestID(ctx context.Context) string {
	return contextString(ctx, RequestIDKey)
}

// ContextFields 从上下文提取日志公共字段。
func ContextFields(ctx context.Context) []Field {
	if ctx == nil {
		return nil
	}
	fields := make([]Field, 0, 2)
	// trace_id 和 request_id 是框架最常用的链路字段，内置提取可以减少业务重复传参。
	if traceID := TraceID(ctx); traceID != "" {
		fields = append(fields, String(string(TraceIDKey), traceID))
	}
	if requestID := RequestID(ctx); requestID != "" {
		fields = append(fields, String(string(RequestIDKey), requestID))
	}
	return fields
}

// contextString 从上下文读取字符串值。
func contextString(ctx context.Context, key ContextKey) string {
	if ctx == nil {
		return ""
	}
	val, _ := ctx.Value(key).(string)
	return val
}
