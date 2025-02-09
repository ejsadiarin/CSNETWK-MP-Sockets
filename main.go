package main

import (
	"fmt"
	"log"
	"net"
)

func main() {
	s := "ejsadiarin"
	fmt.Printf("eeeeeeeeeeee: %s\n", s)
}

func client() {
	conn, err := net.Dial("tcp", "localhost:6969")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	s := "ejsadiarin"
	fmt.Printf("eeeeeeeeeeee: %s\n", s)
}

func server() {
	listener, err := net.Listen("tcp", ":6969")
	if err != nil {
		// slog.Info("server listening on port 6969")
		log.Fatal(err)
	}

	fmt.Println("server listening on port 6969")
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	// create buffer
	buf := make([]byte, 1024) // 1 KB (for 1MB: 1024 * 1024)
	// read data from conn and store it in buf
	n, err := conn.Read(buf)
	if err != nil {
		log.Fatal(err)
	}
	// some processing...
	fmt.Printf("Received data: %v\n", buf[:n])
}
