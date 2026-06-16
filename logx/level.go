package logx

// Level 表示日志输出级别。
type Level string

const (
	// DebugLevel 表示调试级别，适合输出排查细节。
	DebugLevel Level = "debug"
	// InfoLevel 表示信息级别，适合输出关键流程节点。
	InfoLevel Level = "info"
	// WarnLevel 表示警告级别，适合输出可恢复异常。
	WarnLevel Level = "warn"
	// ErrorLevel 表示错误级别，适合输出需要关注的失败。
	ErrorLevel Level = "error"
)
