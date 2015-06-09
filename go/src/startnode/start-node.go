package main

import (
	"flag"
	"log"
	"node"
)

func usage() {
	log.Fatal("Usage: start-node -host <addr> -port <port>")
}

func main() {

	// Set file/line info in logs
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	/**
	Arguments:
		-host  <addr to listen on>
		-port  <port to listen on>
	*/
	var host string
	flag.StringVar(&host, "host", "localhost",
		"Host or local address to listen on")
	port := flag.Int("port", 1234, "Port Number to listen on")

	help := flag.Bool("h", false, "Print Usage and exit")
	helpl := flag.Bool("help", false, "Print Usage and exit")
	flag.Parse()

	if *help || *helpl {
		usage()
	}
	n, err := node.NewNode(host, *port)
	if err != nil {
		log.Fatal(err)
	}
	log.Fatal(n.Start())
}
