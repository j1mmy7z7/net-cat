package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	net "netcat/connections"

)

func main() {
	defaultPort := "8989"

	port := defaultPort
	if len(os.Args) > 1 {
		port = os.Args[1]
	}

	if _, err := strconv.Atoi(port); err != nil {
		fmt.Printf("Invalid port number: %v", err)
	}


	log.Printf("Starting server on port %s...", port)
	server := net.NewServer(":" + port)
	defer close(server.Msgch)
	defer close(server.Quit)
	defer net.CloseLogFile()
	server.Start()
}
