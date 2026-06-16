package logx

import (
	"time"
)

// Field 表示一条结构化日志字段。
type Field struct {
	Key   string // 字段名称
	Value any    // 字段值
}

// String 创建字符串日志字段。
func String(key string, val string) Field {
	return Field{Key: key, Value: val}
}

// Int 创建整数日志字段。
func Int(key string, val int) Field {
	return Field{Key: key, Value: val}
}

// Int64 创建 64 位整数日志字段。
func Int64(key string, val int64) Field {
	return Field{Key: key, Value: val}
}

// Float64 创建浮点数日志字段。
func Float64(key string, val float64) Field {
	return Field{Key: key, Value: val}
}

// Bool 创建布尔日志字段。
func Bool(key string, val bool) Field {
	return Field{Key: key, Value: val}
}

// Any 创建任意类型日志字段。
func Any(key string, val any) Field {
	return Field{Key: key, Value: val}
}

// Err 创建错误日志字段。
func Err(err error) Field {
	return Field{Key: "error", Value: err}
}

// Duration 创建耗时日志字段。
func Duration(key string, val time.Duration) Field {
	return Field{Key: key, Value: val}
}

// Time 创建时间日志字段。
func Time(key string, val time.Time) Field {
	return Field{Key: key, Value: val}
}
