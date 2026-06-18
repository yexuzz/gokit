package logx

import (
	"errors"
	"testing"
)

// TestFieldConstructors 验证常用字段构造器会保留字段名和值。
func TestFieldConstructors(t *testing.T) {
	err := errors.New("failed")
	fields := []Field{
		String("string", "value"),
		Strings("strings", []string{"a", "b"}),
		Int("int", 1),
		Ints("ints", []int{1, 2}),
		Int8("int8", 1),
		Int16("int16", 1),
		Int32("int32", 1),
		Int64("int64", 1),
		Int64s("int64s", []int64{1, 2}),
		Uint("uint", 1),
		Uint8("uint8", 1),
		Uint16("uint16", 1),
		Uint32("uint32", 1),
		Uint64("uint64", 1),
		Uint64s("uint64s", []uint64{1, 2}),
		Uintptr("uintptr", 1),
		Float32("float32", 1.2),
		Float64("float64", 1.2),
		Float64s("float64s", []float64{1.2, 3.4}),
		Bool("bool", true),
		Bools("bools", []bool{true, false}),
		Any("any", map[string]int{"a": 1}),
		Err(err),
		NamedErr("cause", err),
	}

	for _, field := range fields {
		if field.Key == "" {
			t.Fatalf("field key should not be empty: %#v", field)
		}
	}
	defaultErr := fields[len(fields)-2]
	if defaultErr.Key != "error" || defaultErr.Value != err {
		t.Fatalf("want default error field, got %#v", defaultErr)
	}
	namedErr := fields[len(fields)-1]
	if namedErr.Key != "cause" || namedErr.Value != err {
		t.Fatalf("want named error field, got %#v", namedErr)
	}
	strings, ok := fields[1].Value.([]string)
	if !ok || len(strings) != 2 || strings[0] != "a" || strings[1] != "b" {
		t.Fatalf("want strings field, got %#v", fields[1])
	}
}
