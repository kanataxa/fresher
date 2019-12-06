package main

import (
	"log"

	"github.com/kanataxa/fresher"
)

func main() {
	fr := fresher.New()
	if err := fr.Watch(); err != nil {
		log.Fatal(err)
	}
}
