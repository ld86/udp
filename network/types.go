package network

import (
	"crypto/rand"
	"log"
)

type NetworkID [20]byte

func NewNetworkID() NetworkID {
	var id NetworkID
	_, err := rand.Read(id[:])
	if err != nil {
		log.Panicf("rand.Read failed, %s", err)
	}
	return id
}
