package node

import (
	"fmt"

	"github.com/ld86/udp/network"
)

type Node struct {
	Network *network.Network
}

func NewNode() *Node {
	return &Node{
		Network: network.NewNetwork(),
	}
}

func (node *Node) Serve() {
	go node.Network.Serve()

	for {
		select {
		case incomeMessage := <-node.Network.Receive():
			fmt.Println(incomeMessage)
		}
	}
}
