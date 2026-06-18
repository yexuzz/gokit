package logx

import (
	"fmt"

	"go.uber.org/zap"
)

// normalizeLogFields 将上下文字段列表转换为 zap 字段列表。
func normalizeLogFields(fields []Field) []zap.Field {
	if len(fields) == 0 {
		return nil
	}
	zapFields := make([]zap.Field, 0, len(fields))
	for _, field := range fields {
		zapFields = append(zapFields, zap.Any(field.Key, field.Value))
	}
	return zapFields
}

// normalizeZapFields 将业务传入的 logx 字段转换为 zap 字段。
func normalizeZapFields(fields ...Field) []zap.Field {
	if len(fields) == 0 {
		return nil
	}
	zapFields := make([]zap.Field, 0, len(fields))
	for _, field := range fields {
		if field.Key == lineColorFieldKey {
			// 单条日志颜色只写入内部元数据，由 console encoder 使用并从结构化日志中过滤。
			zapFields = append(zapFields, zap.String(lineColorFieldKey, fmt.Sprint(field.Value)))
			continue
		}
		if field.Key == "error" {
			if err, ok := field.Value.(error); ok {
				zapFields = append(zapFields, zap.Error(err))
				continue
			}
		}
		zapFields = append(zapFields, zap.Any(field.Key, field.Value))
	}
	return zapFields
}
