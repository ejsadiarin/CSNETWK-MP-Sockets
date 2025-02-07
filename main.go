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
	c, err := net.Dial("tcp", "localhost:6969")
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()
	s := "ejsadiarin"
	fmt.Printf("eeeeeeeeeeee: %s\n", s)
}

func server() {
	server, err := net.Listen("tcp", ":6969")
	if err != nil {
		// slog.Info("server listening on port 6969")
		log.Fatal(err)
	}

	fmt.Println("server listening on port 6969")
	defer server.Close()

	for {
		server.Accept()
		go handleConnection()
	}
}

func handleConnection() {
}
