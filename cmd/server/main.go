package main

import (
	"log"
	"os"

	"github.com/jinzhu/configor"
)

func main() {
	log.SetOutput(os.Stdout)

	cfg, err := configor.Load("config.yml")
}
