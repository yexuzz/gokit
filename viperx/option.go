package viperx

import "strings"

const (
	// DefaultPath 表示调用方未指定配置文件路径时使用的默认路径.
	DefaultPath = "./config/config.yaml"
)

// Option 用于调整 Viper 初始化行为.
type Option func(*options)

type options struct {
	path           string
	format         Format
	watch          bool
	enableEnv      bool
	envPrefix      string
	envKeyReplacer *strings.Replacer
}

// WithPath 设置配置文件路径.
func WithPath(path string) Option {
	return func(opts *options) {
		opts.path = path
	}
}

// WithFormat 设置配置文件格式, 不设置时根据文件后缀自动识别.
func WithFormat(format Format) Option {
	return func(opts *options) {
		opts.format = format
	}
}

// WithWatch 设置是否监听配置文件变化.
func WithWatch(enable bool) Option {
	return func(opts *options) {
		opts.watch = enable
	}
}

// WithEnv 设置是否启用环境变量覆盖.
func WithEnv(enable bool) Option {
	return func(opts *options) {
		opts.enableEnv = enable
	}
}

// WithEnvPrefix 设置环境变量前缀, 设置后会自动启用环境变量覆盖.
func WithEnvPrefix(prefix string) Option {
	return func(opts *options) {
		opts.enableEnv = true
		opts.envPrefix = prefix
	}
}

// WithEnvKeyReplacer 设置环境变量 key 替换规则, 设置后会自动启用环境变量覆盖.
func WithEnvKeyReplacer(replacer *strings.Replacer) Option {
	return func(opts *options) {
		opts.enableEnv = true
		opts.envKeyReplacer = replacer
	}
}

func defaultOptions(opts ...Option) options {
	cfg := options{
		path:           DefaultPath,
		format:         FormatAuto,
		envKeyReplacer: strings.NewReplacer(".", "_"),
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	return cfg
}
