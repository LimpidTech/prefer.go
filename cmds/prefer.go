package main

import (
	"flag"
	"log"

	"github.com/monokrome/prefer.go"
)

type Configuration struct{}

func init() {
	flag.Parse()
}

func main() {
	data := Configuration{}
	arguments := flag.Args()

	if flag.NArg() != 1 {
		log.Fatalln("Command takes one (and only one) argument.")
	}

	configuration, err := prefer.Load(arguments[0], data)

	if err != nil {
		log.Fatalln(err)
	}

	log.Println("Loaded configuration at", configuration.Identifier)
}
