package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"runtime"
	"strings"
)

func main() {
	fmt.Printf("Number of CPUs: %d\n", runtime.NumCPU())
	fmt.Printf("GOMAXPROCS: %d\n", runtime.GOMAXPROCS(0))
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("What to start (1 - client, 2 - server): ")
	s, err := reader.ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}
	choice := strings.TrimSpace(s)
	switch choice {
	case "1":
		client()
	case "2":
		server()
	default:
		fmt.Print("Invalid choice.")
	}
}

func client() {
	ip, err := net.ResolveTCPAddr("tcp", "localhost:6969")
	if err != nil {
		log.Fatal(err)
	}
	conn, err := net.DialTCP("tcp", nil, ip)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	fmt.Printf("\\? - to list all the available commands\n\n")
	for {
		fmt.Printf("[%s]> ", conn.LocalAddr().String())
		reader := bufio.NewReader(os.Stdin)
		input, err := reader.ReadString('\n')
		msg := strings.TrimSpace(input)
		if err != nil {
			log.Fatal(err)
		}
		if msg == "\\?" {
			fmt.Println("\\join <server_ip> <port> - joins a server given <server_ip> <port>")
			fmt.Println("\\leave - disconnect to server")
			fmt.Println("\\register - register unique handle or alias")
			fmt.Println("\\dir - lists all files from server")
			fmt.Println("\\get - download file from server")
			fmt.Println("\\store - upload file to server")
			fmt.Println("\\? - see all commands")
		} else if strings.Contains(msg, "\\join") {
			if len(strings.Split(msg, " ")) < 3 {
				fmt.Println("\\join needs 2 parameters like so: \\join <server_ip> <port>")
				// TODO: client should not disconnect
				return
			}
			join(conn, msg)
		} else if msg == "\\register" {
		} else if msg == "\\dir" {
		} else if msg == "\\get" {
		} else if msg == "\\store" {
		} else if msg == "\\leave" {
		} else {
			fmt.Println("Invalid command. \\? to see all available commands")
		}
		conn.Write([]byte(msg)) // send what command to server for parsing and logging
	}
}

func join(conn net.Conn, msg string) {
	// parse server_ip and port
	ip := strings.Split(msg, " ")[1]
	port := strings.Split(msg, " ")[2]
	fmt.Println("client: ", ip)
	fmt.Println("client: ", port)
	conn.Write([]byte(ip))
	conn.Write([]byte(port))
}

func server() {
	addr, err := net.ResolveTCPAddr("tcp", ":6969")
	if err != nil {
		// slog.Info("server listening on port 6969")
		log.Fatal(err)
	}
	listener, err := net.ListenTCP("tcp", addr)
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
	for {
		n, err := conn.Read(buf)
		if err != nil {
			if err == io.EOF {
				fmt.Printf("Client: [%s] disconnected\n", conn.RemoteAddr().String())
				log.Printf("Client: [%s] disconnected\n", conn.RemoteAddr().String())
				return
			}
			log.Printf("Error reading from connection: %v\n", err)
		}
		// some processing...
		// TODO: logging: "Received <command> from %s", conn.RemoteAddr().String()
		fmt.Printf("Received command: %v from %s\n", string(buf[:n]), conn.RemoteAddr().String())
		// TODO: logic: parse the commands with their features/capabilities
	}
}
