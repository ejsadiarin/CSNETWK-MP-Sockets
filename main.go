package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"log/slog"
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

// this just initializes the shell
func client() {
	fmt.Printf("/? - to list all the available commands\n\n")
	for {
		fmt.Printf("> ")
		reader := bufio.NewReader(os.Stdin)
		input, err := reader.ReadString('\n')
		msg := strings.TrimSpace(input)
		if err != nil {
			log.Fatal(err)
		}
		if msg == "/?" {
			fmt.Println("/join <server_ip> <port> - joins a server given <server_ip> <port>")
			fmt.Println("/leave - disconnect to server")
			fmt.Println("/register <handle> - register unique handle or alias")
			fmt.Println("/dir - lists all files from server")
			fmt.Println("/get - download file from server")
			fmt.Println("/store - upload file to server")
			fmt.Println("/? - see all commands")
		} else if strings.Contains(msg, "/join") {
			if len(strings.Split(msg, " ")) < 3 {
				fmt.Println("/join needs 2 parameters like so: /join <server_ip> <port>")
				continue
			}
			join(msg)
			// TODO: maybe put the other commands below inside the join func then pass the net.Conn inside there
			// TODO: if match to other "valid" commands (but without joining first) then custom error
		} else {
			fmt.Println("Invalid command. /? to see all available commands")
		}
	}
}

func join(msg string) {
	// parse server_ip and port
	server_ip := strings.Split(msg, " ")[1]
	port := strings.Split(msg, " ")[2]

	ip, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%s", server_ip, port))
	if err != nil {
		log.Fatal(err)
	}

	conn, err := net.DialTCP("tcp", nil, ip)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	fmt.Printf("Successfully connected to server: %s:%s\n", server_ip, port) // for client feedback

	conn.Write([]byte(msg)) // send what command to server for parsing and logging
	for {
		fmt.Printf("> ")
		input, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			log.Println(err)
		}
		msg := strings.TrimSpace(input)

		// TODO: should register first before can access other commands
		if strings.Contains(msg, "/register") {
			if len(strings.Split(msg, " ")) < 2 {
				fmt.Println("/register needs 1 parameter like so: /register Bob")
				continue
			}
			register(conn, msg, strings.Split(msg, " ")[1]) // TODO: test this
		} else if msg == "/dir" || msg == "/get" || msg == "store" {
			fmt.Println("You must register a handle first (e.g. /register Bob)")
		} else if msg == "/leave" {
			// TODO: should be able to leave
		} else if msg == "/join" {
			fmt.Println("Already in connected in a server. You must /leave first before joining another connection.")
		} else if msg == "/?" {
			fmt.Println("/join <server_ip> <port> - joins a server given <server_ip> <port>")
			fmt.Println("/leave - disconnect to server")
			fmt.Println("/register <handle> - register unique handle or alias")
			fmt.Println("/dir - lists all files from server")
			fmt.Println("/get - download file from server")
			fmt.Println("/store - upload file to server")
			fmt.Println("/? - see all commands")
		} else {
			fmt.Println("Invalid command. /? to see all available commands")
		}
	}
}

func register(conn net.Conn, msg string, handle string) {
	fmt.Printf("Successfully registered as %s\n", handle)

	for {
		fmt.Printf("> ")
		input, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			log.Println(err)
		}
		msg := strings.TrimSpace(input)

		if msg == "/dir" {
		} else if msg == "/get" {
		} else if msg == "/store" {
		} else if msg == "/leave" {
			// TODO: should be able to leave
		} else if msg == "/join" {
			fmt.Println("Already in connected in a server. You must /leave first before joining another connection.")
		} else if msg == "/?" {
			fmt.Println("/join <server_ip> <port> - joins a server given <server_ip> <port>")
			fmt.Println("/leave - disconnect to server")
			fmt.Println("/register - register unique handle or alias")
			fmt.Println("/dir - lists all files from server")
			fmt.Println("/get - download file from server")
			fmt.Println("/store - upload file to server")
			fmt.Println("/? - see all commands")
		} else {
			fmt.Println("Invalid command. /? to see all available commands")
		}
	}
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
				slog.Info(fmt.Sprintf("Client: [%s] disconnected\n", conn.RemoteAddr().String()))
				return
			}
			log.Printf("Error reading from connection: %v\n", err)
		}
		// some processing...
		slog.Info(fmt.Sprintf("[%s]: %v\n", conn.RemoteAddr().String(), string(buf[:n]))) // log commands to server
	}
}
