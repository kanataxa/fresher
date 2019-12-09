package fresher

import (
	"fmt"
	"io/ioutil"
	"time"

	yaml "github.com/goccy/go-yaml"
)

type Command struct {
	Name string   `yaml:"name"`
	Args []string `yaml:"arg"`
}

type Config struct {
	Command     *Command       `yaml:"command"`
	Paths       []*WatcherPath `yaml:"path"`
	ExcludePath *GlobalExclude `yaml:"exclude"`
	Extensions  Extensions     `yaml:"extension"`
	Interval    time.Duration  `yaml:"interval"`
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
	if c.Command != nil {
		funcs = append(funcs, ExecCommand(c.Command))
	}
	if len(c.Paths) > 0 {
		funcs = append(funcs, WatchPaths(c.Paths))
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
