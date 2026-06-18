package logx

import (
	"time"

	"github.com/yexuzz/gokit/logx/color"
)

// Field 表示一条结构化日志字段。
type Field struct {
	Key   string // 字段名称
	Value any    // 字段值
}

// LineColor 设置单条日志的终端整行颜色，不会写入 JSON 或文件日志。
func LineColor(ansi color.ANSI) Field {
	return Field{Key: lineColorFieldKey, Value: ansi}
}

// Err 创建错误日志字段。
func Err(err error) Field {
	return Field{Key: "error", Value: err}
}

// String 创建字符串日志字段。
func String(key string, val string) Field {
	return Field{Key: key, Value: val}
}

// Strings 创建字符串切片日志字段。
func Strings(key string, val []string) Field {
	return Field{Key: key, Value: val}
}

// Int 创建整数日志字段。
func Int(key string, val int) Field {
	return Field{Key: key, Value: val}
}

// Ints 创建整数切片日志字段。
func Ints(key string, val []int) Field {
	return Field{Key: key, Value: val}
}

// Int8 创建 8 位整数日志字段。
func Int8(key string, val int8) Field {
	return Field{Key: key, Value: val}
}

// Int16 创建 16 位整数日志字段。
func Int16(key string, val int16) Field {
	return Field{Key: key, Value: val}
}

// Int32 创建 32 位整数日志字段。
func Int32(key string, val int32) Field {
	return Field{Key: key, Value: val}
}

// Int64 创建 64 位整数日志字段。
func Int64(key string, val int64) Field {
	return Field{Key: key, Value: val}
}

// Int64s 创建 64 位整数切片日志字段。
func Int64s(key string, val []int64) Field {
	return Field{Key: key, Value: val}
}

// Uint 创建无符号整数日志字段。
func Uint(key string, val uint) Field {
	return Field{Key: key, Value: val}
}

// Uint8 创建 8 位无符号整数日志字段。
func Uint8(key string, val uint8) Field {
	return Field{Key: key, Value: val}
}

// Uint16 创建 16 位无符号整数日志字段。
func Uint16(key string, val uint16) Field {
	return Field{Key: key, Value: val}
}

// Uint32 创建 32 位无符号整数日志字段。
func Uint32(key string, val uint32) Field {
	return Field{Key: key, Value: val}
}

// Uint64 创建 64 位无符号整数日志字段。
func Uint64(key string, val uint64) Field {
	return Field{Key: key, Value: val}
}

// Uint64s 创建 64 位无符号整数切片日志字段。
func Uint64s(key string, val []uint64) Field {
	return Field{Key: key, Value: val}
}

// Uintptr 创建指针数值日志字段。
func Uintptr(key string, val uintptr) Field {
	return Field{Key: key, Value: val}
}

// Float32 创建 32 位浮点数日志字段。
func Float32(key string, val float32) Field {
	return Field{Key: key, Value: val}
}

// Float64 创建浮点数日志字段。
func Float64(key string, val float64) Field {
	return Field{Key: key, Value: val}
}

// Float64s 创建浮点数切片日志字段。
func Float64s(key string, val []float64) Field {
	return Field{Key: key, Value: val}
}

// Bool 创建布尔日志字段。
func Bool(key string, val bool) Field {
	return Field{Key: key, Value: val}
}

// Bools 创建布尔切片日志字段。
func Bools(key string, val []bool) Field {
	return Field{Key: key, Value: val}
}

// Any 创建任意类型日志字段。
func Any(key string, val any) Field {
	return Field{Key: key, Value: val}
}

// NamedErr 创建自定义字段名的错误日志字段。
func NamedErr(key string, err error) Field {
	return Field{Key: key, Value: err}
}

// Duration 创建耗时日志字段。
func Duration(key string, val time.Duration) Field {
	return Field{Key: key, Value: val}
}

// Time 创建时间日志字段。
func Time(key string, val time.Time) Field {
	return Field{Key: key, Value: val}
}
