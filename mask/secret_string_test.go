package mask

import "testing"

// TestLeft 验证 Left 会替换左侧指定字符数.
func TestLeft(t *testing.T) {
	got := SecretString("123456789012").Left(4, "*")
	if got != "****56789012" {
		t.Fatalf("unexpected masked value: %s", got)
	}
}

// TestLeftShortValue 验证 Left 在字符串长度不足时会全部替换.
func TestLeftShortValue(t *testing.T) {
	got := SecretString("abc").Left(4, "*")
	if got != "***" {
		t.Fatalf("unexpected masked value: %s", got)
	}
}

// TestMiddle 验证 Middle 会替换中间指定字符数.
func TestMiddle(t *testing.T) {
	got := SecretString("13812345678").Middle(4, "*")
	if got != "138****5678" {
		t.Fatalf("unexpected masked value: %s", got)
	}
}

// TestMiddleShortValue 验证 Middle 在字符串长度不足时会全部替换.
func TestMiddleShortValue(t *testing.T) {
	got := SecretString("abc").Middle(4, "*")
	if got != "***" {
		t.Fatalf("unexpected masked value: %s", got)
	}
}

// TestRight 验证 Right 会替换右侧指定字符数.
func TestRight(t *testing.T) {
	got := SecretString("123456789012").Right(4, "*")
	if got != "12345678****" {
		t.Fatalf("unexpected masked value: %s", got)
	}
}

// TestRightShortValue 验证 Right 在字符串长度不足时会全部替换.
func TestRightShortValue(t *testing.T) {
	got := SecretString("abc").Right(4, "*")
	if got != "***" {
		t.Fatalf("unexpected masked value: %s", got)
	}
}

// TestUnicode 验证脱敏按 rune 处理, 不会切坏多字节字符.
func TestUnicode(t *testing.T) {
	left := SecretString("你好世界abc").Left(2, "x")
	if left != "xx世界abc" {
		t.Fatalf("unexpected left masked value: %s", left)
	}
	middle := SecretString("你好世界abc").Middle(4, "x")
	if middle != "你xxxxbc" {
		t.Fatalf("unexpected middle masked value: %s", middle)
	}
	right := SecretString("你好世界abc").Right(2, "x")
	if right != "你好世界axx" {
		t.Fatalf("unexpected right masked value: %s", right)
	}
}
