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
			fmt.Println("/quit - exit application")
			fmt.Println("/? - see all commands")
		} else if strings.Contains(msg, "/join") {
			if len(strings.Split(msg, " ")) < 3 {
				fmt.Println("/join needs 2 parameters like so: /join <server_ip> <port>")
				continue
			}
			quit := join(msg)
			if quit {
				return
			}
			continue
		} else if strings.Contains(msg, "/leave") {
			fmt.Println("You are not yet connected to a server.")
			continue
		} else if strings.Contains(msg, "/dir") || strings.Contains(msg, "/store") || strings.Contains(msg, "/get") || strings.Contains(msg, "/register") {
			fmt.Println("You must join a server first like so: /join localhost 6969")
			continue
		} else if strings.Contains(msg, "/quit") {
			fmt.Println("Quitting application...")
			return
		} else {
			fmt.Println("Invalid command. /? to see all available commands")
			continue
		}
	}
}

func join(msg string) bool {
	// parse server_ip and port
	server_ip := strings.Split(msg, " ")[1]
	port := strings.Split(msg, " ")[2]

	ip, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%s", server_ip, port))
	if err != nil {
		log.Fatal(err)
		// TODO: return false here so app won't crash if err on connecting to server
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
			if len(strings.Split(msg, " ")) < 2 || len(strings.Split(msg, " ")) > 2 {
				fmt.Println("/register has 1 parameter like so: /register Bob")
				continue
			}
			quit := register(conn, msg, strings.Split(msg, " ")[1])
			if quit {
				return true
			}
			break // for /leave: leaving when registered means disconnecting from server
		} else if strings.Contains(msg, "/dir") || strings.Contains(msg, "/get") || strings.Contains(msg, "/store") {
			fmt.Println("You must register a handle first (e.g. /register Bob)")
		} else if strings.Contains(msg, "/leave") {
			fmt.Println("Successfully disconnected from the server!")
			break
		} else if strings.Contains(msg, "/join") {
			fmt.Println("Already in connected in a server. You must /leave first before joining another connection.")
		} else if msg == "/quit" {
			fmt.Println("Quitting application...")
			return true
		} else if msg == "/?" {
			fmt.Println("/join <server_ip> <port> - joins a server given <server_ip> <port>")
			fmt.Println("/leave - disconnect to server")
			fmt.Println("/register <handle> - register unique handle or alias")
			fmt.Println("/dir - lists all files from server")
			fmt.Println("/get - download file from server")
			fmt.Println("/store - upload file to server")
			fmt.Println("/quit - quit application")
			fmt.Println("/? - see all commands")
		} else {
			fmt.Println("Invalid command. /? to see all available commands")
		}
	}
	return false
}

func register(conn net.Conn, msg string, handle string) bool {
	fmt.Printf("Successfully registered as %s\n", handle)

	for {
		fmt.Printf("[%s]> ", handle)
		input, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			log.Println(err)
		}
		msg := strings.TrimSpace(input)

		if msg == "/dir" {
		} else if msg == "/get" {
		} else if msg == "/store" {
		} else if msg == "/leave" {
			fmt.Println("Successfully disconnected from the server!")
			return false
		} else if msg == "/join" {
			fmt.Println("Already in connected in a server. You must /leave first before joining another connection.")
		} else if msg == "/?" {
			fmt.Println("/join <server_ip> <port> - joins a server given <server_ip> <port>")
			fmt.Println("/leave - disconnect to server")
			fmt.Println("/register - register unique handle or alias")
			fmt.Println("/dir - lists all files from server")
			fmt.Println("/get - download file from server")
			fmt.Println("/store - upload file to server")
			fmt.Println("/quit - quit application")
			fmt.Println("/? - see all commands")
		} else if msg == "/quit" {
			fmt.Println("Quitting application...")
			return true // quit signal to be used on the caller function
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
				slog.Info(fmt.Sprintf("Client: [%s] disconnected\n", conn.RemoteAddr().String()))
				return
			}
			log.Printf("Error reading from connection: %v\n", err)
		}
		// some processing...
		slog.Info(fmt.Sprintf("[%s]: %v\n", conn.RemoteAddr().String(), string(buf[:n]))) // log commands to server
	}
}
