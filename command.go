package fresher

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/goccy/go-yaml"
)

type Command struct {
	Name    string
	Arg     []string
	Environ []string
	IsAsync bool
}

func (c *Command) UnmarshalYAML(b []byte) error {
	st := struct {
		Name    string   `yaml:"name"`
		Arg     []string `yaml:"arg"`
		Environ []string `yaml:"env"`
		IsAsync bool     `yaml:"async"`
	}{}
	if err := yaml.Unmarshal(b, &st); err != nil {
		var command string
		if err := yaml.Unmarshal(b, &command); err != nil {
			return err
		}
		words := strings.Split(command, " ")
		for _, word := range words {
			if word == "" {
				continue
			}
			if c.Name == "" {
				c.Name = word
			} else {
				c.Arg = append(c.Arg, word)
			}
		}
		return nil
	}

	c.Name = st.Name
	c.Arg = st.Arg
	c.Environ = st.Environ
	c.IsAsync = st.IsAsync
	return nil
}

func (c *Command) build() *exec.Cmd {
	cmd := exec.Command(c.Name, c.Arg...)
	cmd.Env = c.Environ
	return cmd
}

func (c *Command) Exec() (*os.Process, error) {
	if !c.IsAsync {
		if err := c.runSync(); err != nil {
			return nil, err
		}
		return nil, nil
	}
	p, err := c.runAsync()
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (c *Command) runSync() error {
	cmd := c.build()
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	if _, err := io.Copy(os.Stdout, stdout); err != nil {
		return err
	}
	errBuf, err := ioutil.ReadAll(stderr)
	if err != nil {
		return err
	}
	if len(errBuf) > 0 {
		log.Error(string(errBuf))
		return fmt.Errorf("failed to build: [%s]", string(errBuf))
	}
	return nil
}

func (c *Command) runAsync() (*os.Process, error) {
	cmd := c.build()
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	log.Info("Waiting...")
	if err := cmd.Start(); err != nil {
		return nil, err
	}

	go io.Copy(os.Stdout, stdout)
	go io.Copy(os.Stderr, stderr)

	return cmd.Process, nil
}
