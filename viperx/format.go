package viperx

// Format 表示配置文件的序列化格式.
type Format string

const (
	// FormatAuto 表示根据配置文件后缀自动识别格式.
	FormatAuto Format = ""
	// FormatYAML 表示 YAML 配置格式.
	FormatYAML Format = "yaml"
	// FormatJSON 表示 JSON 配置格式.
	FormatJSON Format = "json"
	// FormatTOML 表示 TOML 配置格式.
	FormatTOML Format = "toml"
)
