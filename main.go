package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"syscall"
	"time"
)

func SendHi(conn net.PacketConn, host string) {
	remoteAddr, err := net.ResolveUDPAddr("udp", host)
	fmt.Println(err)
	conn.WriteTo([]byte("Hi!"), remoteAddr)
}

func StartServer(conn net.PacketConn) {
	for {
		var b [256]byte
		n, addr, _ := conn.ReadFrom(b[:])
		fmt.Printf("%d %s %s\n", n, addr, strings.Trim(string(b[:]), "\r\n "))
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

func main() {
	conn := createPacketConn()
	fmt.Println(conn.LocalAddr())

	if len(os.Args) > 1 {
		for _, host := range os.Args[1:] {
			SendHi(conn, host)

		}
		reader := bufio.NewReader(os.Stdin)
		host, _ := reader.ReadString('\n')
		host = strings.Trim(host, "\n")
		go func(conn net.PacketConn) {
			for {
				var b [256]byte
				n, addr, _ := conn.ReadFrom(b[:])
				fmt.Printf("%d %s %s\n", n, addr, strings.Trim(string(b[:]), "\r\n "))
			}
		}(conn)
		for {
			fmt.Println(host)
			SendHi(conn, host)
			time.Sleep(1 * time.Second)
		}
	} else {
		StartServer(conn)
	}
}
