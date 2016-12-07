package main

import (
	"flag"
	"log"

	"github.com/LimpidTech/prefer.go"
)

type Configuration struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

func init() {
	flag.Parse()
}

func main() {
	data := Configuration{}
	arguments := flag.Args()

	if flag.NArg() != 1 {
		log.Fatalln("Command takes one (and only one) argument.")
	}

	channel, err := prefer.Watch(arguments[0], &data)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println(<-channel)
}
