package mask

import "strings"

const (
	// DefaultMask 表示默认脱敏替换字符.
	DefaultMask = "*"
)

// SecretString 表示需要脱敏展示的字符串.
type SecretString string

func NewSecretString(s string) SecretString {
	return SecretString(s)
}

// Left 替换左侧最多 count 个字符, 字符数不足时会全部替换.
func (s SecretString) Left(count int, mask string) string {
	return left(string(s), count, mask)
}

// Middle 替换中间最多 count 个字符, 字符数不足时会全部替换.
func (s SecretString) Middle(count int, mask string) string {
	return middle(string(s), count, mask)
}

// Right 替换右侧最多 count 个字符, 字符数不足时会全部替换.
func (s SecretString) Right(count int, mask string) string {
	return right(string(s), count, mask)
}

// left 替换左侧最多 count 个字符, 字符数不足时会全部替换.
func left(value string, count int, mask string) string {
	runes := []rune(value)
	n := len(runes)
	if n == 0 || count <= 0 {
		return value
	}
	if mask == "" {
		mask = DefaultMask
	}
	if count > n {
		count = n
	}
	return strings.Repeat(mask, count) + string(runes[count:])
}

// middle 替换中间最多 count 个字符, 字符数不足时会全部替换.
func middle(value string, count int, mask string) string {
	runes := []rune(value)
	n := len(runes)
	if n == 0 || count <= 0 {
		return value
	}
	if mask == "" {
		mask = DefaultMask
	}
	if count > n {
		count = n
	}
	start := (n - count) / 2
	end := start + count
	return string(runes[:start]) + strings.Repeat(mask, count) + string(runes[end:])
}

// right 替换右侧最多 count 个字符, 字符数不足时会全部替换.
func right(value string, count int, mask string) string {
	runes := []rune(value)
	n := len(runes)
	if n == 0 || count <= 0 {
		return value
	}
	if mask == "" {
		mask = DefaultMask
	}
	if count > n {
		count = n
	}
	start := n - count
	return string(runes[:start]) + strings.Repeat(mask, count)
}
