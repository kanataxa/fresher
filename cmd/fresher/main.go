package main

import (
	"log"

	"github.com/jessevdk/go-flags"
	"github.com/kanataxa/fresher"
)

type Option struct {
	Config string `long:"config" short:"c" default:"fresher.yaml" description:"config yaml file name"`
}

var opts Option

func run() error {
	c, err := fresher.LoadConfig(opts.Config)
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
		}
	}
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
