package fresher

import (
	"github.com/goccy/go-yaml"
)

type hostType int

const (
	HostTypeLocal = iota
	HostTypeDocker
)

const (
	hostNameDocker = "docker"
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
	m := make(map[string]string)
	if err := yaml.Unmarshal(b, &m); err != nil {
		return err
	}
	for host, location := range m {
		h.Type = toHostType(host)
		h.LocationName = location
		break
	}
	return nil
}

func toHostType(host string) hostType {
	switch host {
	case hostNameDocker:
		return HostTypeDocker
	}
	return HostTypeLocal
}

func (h *Host) RunCommand(path string) Executor {
	switch h.Type {
	case HostTypeDocker:
		return &DockerCommand{
			Command: &Command{
				Name: "docker",
				Arg: []string{
					"exec",
					h.LocationName,
					path,
				},
				IsAsync: true,
			},
			binPath: path,
			host:    h,
		}
	default:
		return &Command{
			Name:    path,
			IsAsync: true,
		}
	}

}
