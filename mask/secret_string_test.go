package mask

import "testing"

// TestSecretStringMaskWith 验证 SecretString 支持自定义左右保留和中间替换长度.
func TestSecretStringMaskWith(t *testing.T) {
	got := SecretString("123456789012").MaskWith(4, 5, 3, "=")
	if got != "1234=====012" {
		t.Fatalf("unexpected masked value: %s", got)
	}
}

// TestMaskWithMiddleZero 验证 middle 小于等于 0 时替换左右保留区之间的全部字符.
func TestMaskWithMiddleZero(t *testing.T) {
	got := MaskWith("13812345678", 3, 0, 4, "*")
	if got != "138****5678" {
		t.Fatalf("unexpected masked value: %s", got)
	}
}

// TestMaskWithUnicode 验证脱敏按 rune 处理, 不会切坏多字节字符.
func TestMaskWithUnicode(t *testing.T) {
	got := MaskWith("你好世界abc", 1, 3, 2, "x")
	if got != "你xxxabc" {
		t.Fatalf("unexpected masked value: %s", got)
	}
}

// TestMaskWithShortValue 验证短字符串不会因为左右保留长度过大而被破坏.
func TestMaskWithShortValue(t *testing.T) {
	got := MaskWith("abc", 4, 5, 3, "*")
	if got != "abc" {
		t.Fatalf("unexpected masked value: %s", got)
	}
}

// TestSecretStringString 验证默认 String 方法适合手机号一类的常见展示.
func TestSecretStringString(t *testing.T) {
	got := SecretString("13812345678").String()
	if got != "138****5678" {
		t.Fatalf("unexpected masked value: %s", got)
	}
}
