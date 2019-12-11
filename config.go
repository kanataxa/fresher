package fresher

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	yaml "github.com/goccy/go-yaml"
)

type Config struct {
	Build       *BuildConfig     `yaml:"build"`
	Paths       []*WatcherConfig `yaml:"path"`
	ExcludePath *GlobalExclude   `yaml:"exclude"`
	Extensions  Extensions       `yaml:"extension"`
	Interval    time.Duration    `yaml:"interval"`
}

type BuildConfig struct {
	Target string `yaml:"target"`
	OSType string `yaml:"os"`
	Arch   string `yaml:"arch"`
}

func (bc *BuildConfig) Environ() []string {
	env := []string{}
	if bc.OSType != "" {
		env = append(env, fmt.Sprintf("GOOS=%s", bc.OSType))
	}
	if bc.Arch != "" {
		env = append(env, fmt.Sprintf("GOARCH=%s", bc.Arch))
	}
	return env
}

func (bc *BuildConfig) execFilePath() string {
	name := "fresher_run"
	if bc.OSType == "windows" || runtime.GOOS == "windows" {
		name += ".exe"
	}
	return filepath.Join(os.TempDir(), name)
}

func (bc *BuildConfig) BuildCommand() *exec.Cmd {
	cmd := exec.Command("go", "build", "-o", bc.execFilePath(), bc.Target)
	cmd.Env = append(os.Environ(), bc.Environ()...)
	return cmd
}

func (bc *BuildConfig) RunCommand() *exec.Cmd {
	cmd := exec.Command(bc.execFilePath())
	return cmd
}

func (bc *BuildConfig) UnmarshalYAML(b []byte) error {
	st := struct {
		Target string `yaml:"target"`
		OSType string `yaml:"os"`
		Arch   string `yaml:"arch"`
	}{}
	if err := yaml.Unmarshal(b, &st); err != nil {
		var target string
		if err := yaml.Unmarshal(b, &target); err != nil {
			return err
		}
		bc.Target = target
		return nil
	}
	bc.Target = st.Target
	bc.OSType = st.OSType
	bc.Arch = st.Arch
	return nil
}

func LoadConfig(filename string) (*Config, error) {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("fail to load yaml: %w", err)
	}
	var conf Config
	if err := yaml.Unmarshal(b, &conf); err != nil {
		return nil, fmt.Errorf("fail to unmashal yaml: %w", err)
	}
	return &conf, nil
}

func (c *Config) Options() []OptionFunc {
	if c == nil {
		return nil
	}
	var funcs []OptionFunc
	if c.Build != nil {
		funcs = append(funcs, ExecTarget(c.Build))
	}
	if len(c.Paths) > 0 {
		funcs = append(funcs, WatchConfigs(c.Paths))
	}
	if c.ExcludePath != nil {
		funcs = append(funcs, GlobalExcludePath(c.ExcludePath))
	}
	if len(c.Extensions) > 0 {
		funcs = append(funcs, ExtensionPaths(c.Extensions))
	}
	if c.Interval > 0 {
		funcs = append(funcs, WatchInterval(c.Interval*time.Second))
	}
	return funcs
}
