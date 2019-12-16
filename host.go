package fresher

import (
	"fmt"

	"github.com/goccy/go-yaml"
)

type hostType int

const (
	HostTypeLocal = iota
	HostTypeDocker
)

func (h hostType) String() string {
	switch h {
	case HostTypeDocker:
		return "docker"
	default:
		return "localhost"
	}
}

type Host struct {
	Type         hostType
	LocationName string
}

func (h *Host) UnmarshalYAML(b []byte) error {
	st := struct {
		Docker string `yaml:"docker"`
	}{}
	if err := yaml.Unmarshal(b, &st); err != nil {
		return err
	}
	if st.Docker != "" {
		h.Type = HostTypeDocker
		h.LocationName = st.Docker
	}
	return nil
}

func (h *Host) RunCommands(path string) []Executor {
	switch h.Type {
	case HostTypeDocker:
		return []Executor{
			&Command{
				Name: "docker",
				Arg: []string{
					"cp",
					path,
					fmt.Sprintf("%s:%s", h.LocationName, path),
				},
			},
			&Command{
				Name: "docker",
				Arg: []string{
					"exec",
					h.LocationName,
					path,
				},
				IsAsync: true,
			},
		}
	default:
		return []Executor{
			&Command{
				Name:    path,
				IsAsync: true,
			},
		}
	}

}
