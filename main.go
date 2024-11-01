package main

import (
	"fmt"
	"log"
	"os"

	"strconv"

	net "netcat/connections"
)

func isValidPort(port int) bool {
	return port >= 0 && port <= 65535
}

func main() {
	var port string

	args := os.Args[1:]

	if len(args) > 1 {
		fmt.Println("[USAGE]: ./TCPChat $port")
		return
	}
	if len(args) == 1 {
		port = args[0]
	}
	if len(args) == 0 {
		port = "8989"
	}

	checker, err := strconv.Atoi(port)
	if err != nil {
		fmt.Printf("Invalid port number: %v", err)
		return
	}

	if !isValidPort(checker) {
		log.Fatal("Enter a valid port number between 0 - 65535")

	}

	log.Printf("Starting server on port %s...", port)
	server := net.NewServer(":" + port)
	defer close(server.Msgch)
	defer close(server.Quit)
	defer net.CloseLogFile()
	server.Start()
}
