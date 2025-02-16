package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net"
	"os"
	"path/filepath"
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

	// notify server that client registered
	conn.Write([]byte(fmt.Sprintf("client %s registered as  %s\n", conn.LocalAddr().String(), handle)))

	for {
		fmt.Printf("[%s]> ", handle)
		input, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			log.Println(err)
		}
		msg := strings.TrimSpace(input)
		log.Printf("msg: %v", msg)

		if msg == "/dir" {
			dir() // NOTE: only works locally, client and server should communicate via the connection only
		} else if strings.Contains(msg, "/get") {
			_, filename, found := strings.Cut(msg, " ")
			if !found {
				fmt.Println("missing filename parameter for /get")
				continue
			}

			// NOTE: should request "/get filename" to server via connection (conn.Write) like so:
			request := fmt.Sprintf("/get %s", filename)
			_, err := conn.Write([]byte(request))
			if err != nil {
				fmt.Printf("cannot communicate with the server: %w\n", err)
			}

			bufSize := make([]byte, 8) // represents buffer of type uint64
			// read file size first so we know until when to read without waiting for an EOF signal or error
			_, err = io.ReadFull(conn, bufSize)
			if err != nil {
				fmt.Printf("cannot read file size: %v\n", err)
				continue
			}
			// little endian is default in most network protocols and modern CPUs (x86 and ARM64)
			fileSize := binary.LittleEndian.Uint64(bufSize)

			out, err := os.Create(filename)
			if err != nil {
				fmt.Printf("cannot create file: %v\n", err)
				continue
			}
			defer out.Close()

			// read file contents from connection to out file
			_, err = io.CopyN(out, conn, int64(fileSize))
			if err != nil {
				fmt.Printf("cannot receive file: %w\n", err)
				continue

			}
			fmt.Printf("Successfully downloaded %s (%d bytes)\n", filename, fileSize)
		} else if msg == "/store" {
		} else if msg == "/leave" {
			fmt.Println("Successfully disconnected from the server!")
			return false
		} else if strings.Contains(msg, "/join") {
			fmt.Println("Already in connected in a server. You must /leave first before joining another connection.")
		} else if strings.Contains(msg, "/register") {
			fmt.Printf("You are already registered as %s.\n", handle)
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

func dir() {
	dir, err := os.ReadDir("./dir")
	if err != nil {
		fmt.Println("error reading directory")
	}

	// NOTE: can trim the "-" prefix before the file names
	for _, v := range dir {
		fmt.Println(v)
	}
}

// TODO: implement these features
func store() {
	// upload
}

func getHandler(conn net.Conn, filename string) error {
	filedata, err := os.ReadFile(filepath.Join("dir", filename))
	if err != nil {
		if err == os.ErrNotExist {
			slog.Error(fmt.Sprintf("cannot read file that doesn't exist: %s, err: %v\n", filename, err))
		}
		return fmt.Errorf("file doesn't exist: %v\n", err)
	}

	// send the filesize first for the client to know how big is the file to be created and copied in the client
	bufSize := make([]byte, 8)
	binary.LittleEndian.PutUint64(bufSize, uint64(len(filedata)))
	_, err = conn.Write(bufSize)
	if err != nil {
		return fmt.Errorf("cannot send file size to client: %w\n", err)
	}

	// then send the whole file data
	_, err = conn.Write(filedata)
	if err != nil {
		slog.Error(fmt.Sprintf("cannot read file that doesn't exist: %s, err: %v\n", filename, err))
		return fmt.Errorf("failed to send file size: %v\n", err)
	}

	slog.Info(fmt.Sprintf("sending file: %s to client...", filename))
	return nil
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
		cmd := string(buf[:n])
		slog.Info(fmt.Sprintf("[%s]: %v\n", conn.RemoteAddr().String(), cmd)) // log commands to server
		if strings.HasPrefix(cmd, "/get") {
			if len(strings.Fields(cmd)) < 2 {
				slog.Error("missing filename for /get")
				_, err := conn.Write([]byte("missing filename for /get"))
				if err != nil {
					slog.Error("cannot write response to connection")
				}
				continue
			}
			filename := strings.Fields(cmd)[1]
			if err := getHandler(conn, filename); err != nil {
				slog.Error(fmt.Sprintf("Error handling get: %v\n", err))
			}
		}
	}
}
