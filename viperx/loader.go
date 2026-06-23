package viperx

import (
	"fmt"

	"github.com/spf13/viper"
)

// Load 从包级默认 Viper 实例解析配置到业务结构体.
func Load[T any]() (T, error) {
	return LoadFrom[T](Default())
}

// MustLoad 从包级默认 Viper 实例解析配置到业务结构体, 失败时 panic.
func MustLoad[T any]() T {
	return MustLoadFrom[T](Default())
}

// LoadFrom 从指定 Viper 实例解析配置到业务结构体.
func LoadFrom[T any](vp *viper.Viper) (T, error) {
	var result T
	if vp == nil {
		return result, fmt.Errorf("viperx: nil viper")
	}
	if err := vp.Unmarshal(&result); err != nil {
		return result, fmt.Errorf("viperx: unmarshal config: %w", err)
	}
	return result, nil
}

// MustLoadFrom 从指定 Viper 实例解析配置到业务结构体, 失败时 panic.
func MustLoadFrom[T any](vp *viper.Viper) T {
	cfg, err := LoadFrom[T](vp)
	if err != nil {
		panic(err)
	}
	return cfg
}

// LoadKey 从包级默认 Viper 实例解析指定配置节点到业务结构体.
func LoadKey[T any](key string) (T, error) {
	return LoadKeyFrom[T](Default(), key)
}

// MustLoadKey 从包级默认 Viper 实例解析指定配置节点到业务结构体, 失败时 panic.
func MustLoadKey[T any](key string) T {
	return MustLoadKeyFrom[T](Default(), key)
}

// LoadKeyFrom 从指定 Viper 实例解析指定配置节点到业务结构体.
func LoadKeyFrom[T any](vp *viper.Viper, key string) (T, error) {
	var result T
	if vp == nil {
		return result, fmt.Errorf("viperx: nil viper")
	}
	if err := vp.UnmarshalKey(key, &result); err != nil {
		return result, fmt.Errorf("viperx: unmarshal config key %s: %w", key, err)
	}
	return result, nil
}

// MustLoadKeyFrom 从指定 Viper 实例解析指定配置节点到业务结构体, 失败时 panic.
func MustLoadKeyFrom[T any](vp *viper.Viper, key string) T {
	cfg, err := LoadKeyFrom[T](vp, key)
	if err != nil {
		panic(err)
	}
	return cfg
}
