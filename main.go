package main

import (
	"log"
)

func main() {
	app := new(app)
	log.Fatalf("%+v", app.run())
}
