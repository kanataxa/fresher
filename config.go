package fresher

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/goccy/go-yaml"
)

type Config struct {
	Build       *BuildConfig     `yaml:"build"`
	Paths       []*WatcherConfig `yaml:"path"`
	ExcludePath *GlobalExclude   `yaml:"exclude"`
	Extensions  Extensions       `yaml:"extension"`
	Interval    time.Duration    `yaml:"interval"`
}

type BuildConfig struct {
	Target         string     `yaml:"target"`
	Host           *Host      `yaml:"host"`
	Output         string     `yaml:"output"`
	Environ        Environ    `yaml:"env"`
	Arg            []string   `yaml:"arg"`
	WithoutRun     bool       `yaml:"without_run"`
	BeforeCommands []*Command `yaml:"before"`
	AfterCommands  []*Command `yaml:"after"`
}

func (bc *BuildConfig) runBinaryPath() string {
	if bc.Output != "" {
		return bc.Output
	}
	name := "fresher_run"
	var isIncludeGOOSEnv bool
	for _, env := range bc.Environ {
		if strings.Contains(env, "GOOS") {
			isIncludeGOOSEnv = true
			if strings.Contains(env, "windows") {
				name += ".exe"
			}
			break
		}
	}
	if !isIncludeGOOSEnv && runtime.GOOS == "windows" {
		name += ".exe"
	}
	return filepath.Join(os.TempDir(), name)
}

func (bc *BuildConfig) buildArg() []string {
	arg := []string{"build", "-o", bc.runBinaryPath()}
	if len(bc.Arg) > 0 {
		arg = append(arg, bc.Arg...)
	}
	arg = append(arg, bc.Target)
	return arg
}

func (bc *BuildConfig) Commands() []Executor {
	commands := []Executor{
		bc.BuildCommand(),
	}
	if cmd := bc.RunCommand(); cmd != nil {
		commands = append(commands, cmd)
	}
	if len(bc.BeforeCommands) > 0 {
		cmds := make([]Executor, len(bc.BeforeCommands))
		for idx, cmd := range bc.BeforeCommands {
			cmds[idx] = cmd
		}
		commands = append(cmds, commands...)
	}
	if len(bc.AfterCommands) > 0 {
		cmds := make([]Executor, len(bc.AfterCommands))
		for idx, cmd := range bc.AfterCommands {
			cmds[idx] = cmd
		}
		commands = append(commands, cmds...)
	}
	return commands
}

func (bc *BuildConfig) BuildCommand() Executor {
	cmd := exec.Command("go", bc.buildArg()...)
	cmd.Env = append(os.Environ(), bc.Environ...)
	return &Command{
		Name:    "go",
		Arg:     bc.buildArg(),
		Environ: append(os.Environ(), bc.Environ...),
		IsAsync: false,
	}
}

func (bc *BuildConfig) RunCommand() Executor {
	if bc.WithoutRun {
		return nil
	}
	if bc.Host == nil {
		return &Command{
			Name:    bc.runBinaryPath(),
			IsAsync: true,
		}
	}
	return bc.Host.RunCommand(bc.runBinaryPath())
}

type Environ []string

func (e *Environ) UnmarshalYAML(b []byte) error {
	m := make(map[string]string)
	if err := yaml.Unmarshal(b, &m); err != nil {
		return err
	}
	for k, v := range m {
		*e = append(*e, fmt.Sprintf("%s=%s", k, v))
	}
	return nil
}

func (bc *BuildConfig) UnmarshalYAML(b []byte) error {
	st := struct {
		Target         string     `yaml:"target"`
		Host           *Host      `yaml:"host"`
		Output         string     `yaml:"output"`
		Environ        Environ    `yaml:"env"`
		Arg            []string   `yaml:"arg"`
		WithoutRun     bool       `yaml:"without_run"`
		BeforeCommands []*Command `yaml:"before"`
		AfterCommands  []*Command `yaml:"after"`
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
	bc.Host = st.Host
	bc.Output = st.Output
	bc.Environ = st.Environ
	bc.Arg = st.Arg
	bc.WithoutRun = st.WithoutRun
	bc.BeforeCommands = st.BeforeCommands
	bc.AfterCommands = st.AfterCommands
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
