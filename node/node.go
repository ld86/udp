package node

import (
	"fmt"

	"github.com/ld86/udp/network"
)

type Node struct {
	Network *network.Network
}

func NewNode() (*Node, error) {
	network, err := network.NewNetwork()
	if err != nil {
		return nil, err
	}

	node := &Node{
		Network: network,
	}

	return node, nil
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
