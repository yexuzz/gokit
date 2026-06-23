package viperx

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"github.com/spf13/viper"
)

var (
	defaultMu sync.RWMutex
	defaultV  = viper.New()
)

// Init 初始化包级默认 Viper 实例.
//
// 适合应用启动阶段调用一次, 后续可以通过 Default, Load 和 MustLoad 复用该实例.
func Init(opts ...Option) error {
	vp, err := NewViper(opts...)
	if err != nil {
		return err
	}
	defaultMu.Lock()
	defaultV = vp
	defaultMu.Unlock()
	return nil
}

// MustInit 初始化包级默认 Viper 实例, 失败时 panic.
func MustInit(opts ...Option) {
	if err := Init(opts...); err != nil {
		panic(err)
	}
}

// Default 返回包级默认 Viper 实例.
func Default() *viper.Viper {
	defaultMu.RLock()
	defer defaultMu.RUnlock()
	return defaultV
}

// NewViper 创建并初始化一个独立的 Viper 实例.
func NewViper(opts ...Option) (*viper.Viper, error) {
	cfg := defaultOptions(opts...)
	format, err := resolveFormat(cfg.path, cfg.format)
	if err != nil {
		return nil, err
	}
	vp := viper.New()
	vp.SetConfigFile(cfg.path)
	vp.SetConfigType(string(format))
	if cfg.enableEnv {
		if cfg.envPrefix != "" {
			vp.SetEnvPrefix(cfg.envPrefix)
		}
		if cfg.envKeyReplacer != nil {
			vp.SetEnvKeyReplacer(cfg.envKeyReplacer)
		}
		vp.AutomaticEnv()
	}
	if err = vp.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("viperx: read config %s: %w", cfg.path, err)
	}
	if cfg.watch {
		vp.WatchConfig()
	}
	return vp, nil
}

// MustNewViper 创建并初始化一个独立的 Viper 实例, 失败时 panic.
func MustNewViper(opts ...Option) *viper.Viper {
	vp, err := NewViper(opts...)
	if err != nil {
		panic(err)
	}
	return vp
}

func resolveFormat(path string, format Format) (Format, error) {
	if format != FormatAuto {
		return normalizeFormat(format)
	}
	ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(path)), ".")
	return normalizeFormat(Format(ext))
}

func normalizeFormat(format Format) (Format, error) {
	switch strings.ToLower(string(format)) {
	case "yaml", "yml":
		return FormatYAML, nil
	case "json":
		return FormatJSON, nil
	case "toml":
		return FormatTOML, nil
	default:
		return "", fmt.Errorf("viperx: unsupported config format %q", format)
	}
}
