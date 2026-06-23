package viperx

import (
	"os"
	"path/filepath"
	"testing"
)

type testConfig struct {
	App struct {
		Name string `json:"name" mapstructure:"name" toml:"name" yaml:"name"`
		Env  string `json:"env" mapstructure:"env" toml:"env" yaml:"env"`
	} `json:"app" mapstructure:"app" toml:"app" yaml:"app"`
	Port int `json:"port" mapstructure:"port" toml:"port" yaml:"port"`
}

type appConfig struct {
	Name string `mapstructure:"name"`
	Env  string `mapstructure:"env"`
}

// TestNewViperYAML 验证 YAML 配置可以按文件后缀自动识别并加载.
func TestNewViperYAML(t *testing.T) {
	path := writeTestConfig(t, "config.yaml", []byte(`
app:
  name: gokit
  env: dev
port: 8080
`))
	vp, err := NewViper(WithPath(path))
	if err != nil {
		t.Fatalf("new viper: %v", err)
	}
	cfg, err := LoadFrom[testConfig](vp)
	if err != nil {
		t.Fatalf("load yaml: %v", err)
	}
	if cfg.App.Name != "gokit" || cfg.App.Env != "dev" || cfg.Port != 8080 {
		t.Fatalf("unexpected config: %#v", cfg)
	}
}

// TestNewViperJSON 验证 JSON 配置可以按文件后缀自动识别并加载.
func TestNewViperJSON(t *testing.T) {
	path := writeTestConfig(t, "config.json", []byte(`{"app":{"name":"gokit","env":"prod"},"port":9090}`))
	vp, err := NewViper(WithPath(path))
	if err != nil {
		t.Fatalf("new viper: %v", err)
	}
	cfg, err := LoadFrom[testConfig](vp)
	if err != nil {
		t.Fatalf("load json: %v", err)
	}
	if cfg.App.Name != "gokit" || cfg.App.Env != "prod" || cfg.Port != 9090 {
		t.Fatalf("unexpected config: %#v", cfg)
	}
}

// TestNewViperTOML 验证 TOML 配置可以按文件后缀自动识别并加载.
func TestNewViperTOML(t *testing.T) {
	path := writeTestConfig(t, "config.toml", []byte(`
port = 7070

[app]
name = "gokit"
env = "test"
`))
	vp, err := NewViper(WithPath(path))
	if err != nil {
		t.Fatalf("new viper: %v", err)
	}
	cfg, err := LoadFrom[testConfig](vp)
	if err != nil {
		t.Fatalf("load toml: %v", err)
	}
	if cfg.App.Name != "gokit" || cfg.App.Env != "test" || cfg.Port != 7070 {
		t.Fatalf("unexpected config: %#v", cfg)
	}
}

// TestNewViperWithFormat 验证调用方可以显式指定配置格式.
func TestNewViperWithFormat(t *testing.T) {
	path := writeTestConfig(t, "config.conf", []byte(`{"app":{"name":"gokit","env":"dev"},"port":6060}`))
	vp, err := NewViper(WithPath(path), WithFormat(FormatJSON))
	if err != nil {
		t.Fatalf("new viper: %v", err)
	}
	cfg, err := LoadFrom[testConfig](vp)
	if err != nil {
		t.Fatalf("load explicit format: %v", err)
	}
	if cfg.Port != 6060 {
		t.Fatalf("unexpected port: %d", cfg.Port)
	}
}

// TestNewViperWithEnv 验证启用环境变量后可以覆盖配置文件中的同名配置.
func TestNewViperWithEnv(t *testing.T) {
	t.Setenv("GOKIT_APP_ENV", "prod")
	t.Setenv("GOKIT_PORT", "5050")
	path := writeTestConfig(t, "config.yaml", []byte(`
app:
  name: gokit
  env: dev
port: 8080
`))
	vp, err := NewViper(WithPath(path), WithEnvPrefix("GOKIT"))
	if err != nil {
		t.Fatalf("new viper: %v", err)
	}
	cfg, err := LoadFrom[testConfig](vp)
	if err != nil {
		t.Fatalf("load with env: %v", err)
	}
	if cfg.App.Name != "gokit" || cfg.App.Env != "prod" || cfg.Port != 5050 {
		t.Fatalf("unexpected config: %#v", cfg)
	}
}

// TestDefaultViper 验证包级默认 Viper 可以初始化一次后复用.
func TestDefaultViper(t *testing.T) {
	old := Default()
	t.Cleanup(func() {
		defaultMu.Lock()
		defaultV = old
		defaultMu.Unlock()
	})
	path := writeTestConfig(t, "config.yaml", []byte(`
app:
  name: gokit
  env: dev
port: 8080
`))
	if err := Init(WithPath(path)); err != nil {
		t.Fatalf("init default viper: %v", err)
	}
	cfg, err := Load[testConfig]()
	if err != nil {
		t.Fatalf("load default config: %v", err)
	}
	if cfg.App.Name != "gokit" || cfg.Port != 8080 {
		t.Fatalf("unexpected config: %#v", cfg)
	}
}

// TestLoadKey 验证可以解析指定配置节点.
func TestLoadKey(t *testing.T) {
	path := writeTestConfig(t, "config.yaml", []byte(`
app:
  name: gokit
  env: dev
port: 8080
`))
	vp, err := NewViper(WithPath(path))
	if err != nil {
		t.Fatalf("new viper: %v", err)
	}
	app, err := LoadKeyFrom[appConfig](vp, "app")
	if err != nil {
		t.Fatalf("load key: %v", err)
	}
	if app.Name != "gokit" || app.Env != "dev" {
		t.Fatalf("unexpected app config: %#v", app)
	}
}

// TestLoadKeyString 验证可以解析字符串配置节点.
func TestLoadKeyString(t *testing.T) {
	path := writeTestConfig(t, "config.yaml", []byte(`
app:
  name: gokit
a: s
`))
	vp, err := NewViper(WithPath(path))
	if err != nil {
		t.Fatalf("new viper: %v", err)
	}
	value, err := LoadKeyFrom[string](vp, "a")
	if err != nil {
		t.Fatalf("load string key: %v", err)
	}
	if value != "s" {
		t.Fatalf("unexpected value: %s", value)
	}
}

// TestNewViperUnsupportedFormat 验证未知配置格式会返回明确错误.
func TestNewViperUnsupportedFormat(t *testing.T) {
	path := writeTestConfig(t, "config.ini", []byte(`app.name=gokit`))
	if _, err := NewViper(WithPath(path)); err == nil {
		t.Fatal("expect unsupported format error")
	}
}

func writeTestConfig(t *testing.T, name string, data []byte) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write test config: %v", err)
	}
	return path
}
