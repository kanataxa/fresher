package main

import (
	"os"

	"github.com/jessevdk/go-flags"
	"github.com/kanataxa/fresher"
)

type Option struct {
	Start StartCommand `description:"start fresher and watch files" command:"start"`
}

var opts Option

type StartCommand struct {
	Config string `long:"config" short:"c" default:"fresher.yaml" description:"config yaml file name"`
}

func (s *StartCommand) Execute(args []string) error {
	c, err := fresher.LoadConfig(s.Config)
	if err != nil {
		return err
	}
	fr := fresher.New(c.Options()...)
	if err := fr.Watch(); err != nil {
		return err
	}
	return nil
}

func main() {
	parser := flags.NewParser(&opts, flags.Default)
	if _, err := parser.Parse(); err != nil {
		if e, ok := err.(*flags.Error); ok {
			if e.Type == flags.ErrHelp {
				return
			}
			parser.WriteHelp(os.Stdout)
		}
	}

}
