package network

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"syscall"
)

type OutcomeMessage struct {
	DstAddr string
	Payload []byte
}

type IncomeMessage struct {
	SrcLocalAddr  []string
	SrcGlobalAddr string
	Payload       []byte
}

type Message struct {
	LocalAddr []string
	Payload   []byte
}

type Network struct {
	serverConn      net.PacketConn
	numberOfPackets sync.Map
	localAddr       []string

	sent     chan OutcomeMessage
	received chan IncomeMessage
}

func NewNetwork() (*Network, error) {
	localIPs, err := localIPs()

	if err != nil {
		return nil, err
	}

	network := &Network{
		serverConn: createPacketConn(),
		sent:       make(chan OutcomeMessage),
		received:   make(chan IncomeMessage),
		localAddr:  make([]string, 0),
	}

	port := network.serverConn.LocalAddr().(*net.UDPAddr).Port
	for _, localIP := range localIPs {
		localAddr := fmt.Sprintf("%s:%d", localIP, port)
		network.localAddr = append(network.localAddr, localAddr)
	}

	return network, nil
}

func (network *Network) Send(udpAddr string, payload []byte) {
	outcomeMessage := OutcomeMessage{
		DstAddr: udpAddr,
		Payload: payload,
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
				LocalAddr: network.localAddr,
				Payload:   outcomeMessage.Payload,
			}
			network.marshalAndSend(outcomeMessage.DstAddr, message)
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
			SrcLocalAddr:  message.LocalAddr,
			SrcGlobalAddr: sourceGlobalIP,
			Payload:       message.Payload,
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

func localIPs() ([]string, error) {
	interfaces, err := net.Interfaces()

	result := make([]string, 0)

	if err != nil {
		return nil, err
	}

	for _, i := range interfaces {
		if i.Flags&net.FlagUp == 0 {
			continue
		}
		if i.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, err := i.Addrs()

		if err != nil {
			return nil, err
		}

		for _, addr := range addrs {
			var ip net.IP

			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}

			if ip == nil || ip.IsLoopback() {
				continue
			}

			ip = ip.To4()

			if ip == nil {
				continue
			}

			result = append(result, ip.String())
		}
	}

	return result, nil
}
