package main

import (
	"flag"
)

var inputs = flag.String("in", "udp://0.0.0.0:12221", "comma-separated list of GELF input addresses")
var outputs = flag.String("out", "tcp://127.0.0.0:12221", "comma-separated list of output Graylog servers (TCP or HTTP)")

func init() {
	flag.Parse()
}

func main() {
}
