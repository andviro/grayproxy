package main

import (
	"log"
)

var version string

func main() {
	app := new(app)
	log.Fatalf("%+v", app.run())
}
