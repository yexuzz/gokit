package mask

import "strings"

const (
	// DefaultMask 表示默认脱敏替换字符.
	DefaultMask = "*"
)

// SecretString 表示需要脱敏展示的字符串.
type SecretString string

// Raw 返回未脱敏的原始字符串.
func (s SecretString) Raw() string {
	return string(s)
}

// String 返回默认脱敏展示文本.
func (s SecretString) String() string {
	return s.Mask(3, 0, 4)
}

// Mask 使用默认替换字符脱敏字符串.
//
// left 表示左侧保留字符数, middle 表示中间替换字符数, right 表示右侧保留字符数.
// middle 小于等于 0 时会替换左侧和右侧之间的全部字符.
func (s SecretString) Mask(left int, middle int, right int) string {
	return s.MaskWith(left, middle, right, DefaultMask)
}

// MaskWith 使用指定替换字符脱敏字符串.
//
// 例如 MaskWith(4, 5, 3, "=") 表示保留左侧 4 个字符, 中间 5 个字符替换为 "=", 右侧保留 3 个字符.
func (s SecretString) MaskWith(left int, middle int, right int, mask string) string {
	return MaskWith(string(s), left, middle, right, mask)
}

// Mask 使用默认替换字符脱敏字符串.
func Mask(value string, left int, middle int, right int) string {
	return MaskWith(value, left, middle, right, DefaultMask)
}

// MaskWith 使用指定替换字符脱敏字符串.
//
// left 和 right 表示两侧保留字符数. middle 表示从左侧保留区后开始替换的字符数.
// middle 小于等于 0 时会替换左侧和右侧之间的全部字符.
func MaskWith(value string, left int, middle int, right int, mask string) string {
	runes := []rune(value)
	n := len(runes)
	if n == 0 {
		return ""
	}
	if mask == "" {
		mask = DefaultMask
	}
	if left < 0 {
		left = 0
	}
	if right < 0 {
		right = 0
	}
	if left > n {
		left = n
	}
	rightStart := n - right
	if rightStart < left {
		rightStart = left
	}
	if middle <= 0 || left+middle > rightStart {
		middle = rightStart - left
	}
	if middle <= 0 {
		return value
	}
	maskedEnd := left + middle
	return string(runes[:left]) + strings.Repeat(mask, middle) + string(runes[maskedEnd:])
}
