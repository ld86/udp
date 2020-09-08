package network

import (
	"encoding/json"
	"log"
	"net"
	"os"
	"sync"
	"syscall"
)

type OutcomeMessage struct {
	DestinationIP string
	Payload       []byte
}

type IncomeMessage struct {
	SourceLocalIP  string
	SourceGlobalIP string
	Payload        []byte
}

type Message struct {
	LocalIP string
	Payload []byte
}

type Network struct {
	serverConn      net.PacketConn
	numberOfPackets sync.Map

	sent     chan OutcomeMessage
	received chan IncomeMessage
}

func NewNetwork() *Network {
	return &Network{
		serverConn: createPacketConn(),
		sent:       make(chan OutcomeMessage),
		received:   make(chan IncomeMessage),
	}
}

func (network *Network) Send(udpAddr string, payload []byte) {
	outcomeMessage := OutcomeMessage{
		DestinationIP: udpAddr,
		Payload:       payload,
	}
	network.sent <- outcomeMessage
}

func (network *Network) Receive() chan IncomeMessage {
	return network.received
}

func (network *Network) Serve() {
	log.Println(network.serverConn.LocalAddr())
	go network.handleReceived()
	network.handleSent()
}

func (network *Network) marshalAndSend(udpAddr string, message *Message) {
	remoteAddr, _ := net.ResolveUDPAddr("udp", udpAddr)
	data, _ := json.Marshal(message)
	network.serverConn.WriteTo(data, remoteAddr)
}

func (network *Network) handleSent() {
	for {
		select {
		case outcomeMessage := <-network.sent:
			message := &Message{
				LocalIP: network.serverConn.LocalAddr().String(),
				Payload: outcomeMessage.Payload,
			}
			network.marshalAndSend(outcomeMessage.DestinationIP, message)
		}
	}
}

func (network *Network) handleReceived() {
	for {
		var buffer [2048]byte
		var message Message
		n, remoteAddr, _ := network.serverConn.ReadFrom(buffer[:])
		json.Unmarshal(buffer[:n], &message)

		sourceGlobalIP := remoteAddr.String()

		incomeMessage := IncomeMessage{
			SourceLocalIP:  message.LocalIP,
			SourceGlobalIP: sourceGlobalIP,
			Payload:        message.Payload,
		}

		network.received <- incomeMessage
	}
}

func createPacketConn() net.PacketConn {
	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_DGRAM, syscall.IPPROTO_UDP)

	if err != nil {
		log.Fatalf("Cannot create socket, %s", err)
	}

	if err := syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1); err != nil {
		log.Fatalf("Cannot set SO_REUSEADDR on socket, %s", err)
	}

	udpAddr, err := net.ResolveUDPAddr("udp", ":0")
	if err != nil && udpAddr.IP != nil {
		log.Fatalf("Cannot resolve addr, %s", err)
	}

	if err := syscall.Bind(fd, &syscall.SockaddrInet4{Port: udpAddr.Port}); err != nil {
		log.Fatalf("Cannot bind socket, %s", err)
	}

	file := os.NewFile(uintptr(fd), string(fd))
	conn, err := net.FilePacketConn(file)
	if err != nil {
		log.Fatalf("Cannot create connection from socket, %s", err)
	}

	if err = file.Close(); err != nil {
		log.Fatalf("Cannot close dup file, %s", err)
	}

	return conn
}
