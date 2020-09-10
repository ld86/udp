package main

import (
	"os"

	"github.com/ld86/udp/node"
)

func main() {
	node, err := node.NewNode()

	if err != nil {
		panic(err)
	}

	if len(os.Args) == 1 {
		node.Serve()
	} else {
		go node.Serve()
		remoteIP := os.Args[1]
		node.Network.Send(remoteIP, []byte("Hi!"))
	}
}
