package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	net "netcat/connections"

)

func main() {
	var port string


	Args := os.Args[1:]

	if len(Args) == 0 {
		port = "8989"
	} else if len(Args) == 1 {
		port = Args[0]
	} else {
		fmt.Println("[USAGE]: ./TCPChat $port")
		return
	}

	checker , err := strconv.Atoi(port); 
	if err != nil {
		fmt.Printf("Invalid port number: %v", err)
		return
	}
	
	if checker < 0 || checker > 65535 {
		fmt.Println("Enter a valid port number between 0 - 65535")
		return
	}


	log.Printf("Starting server on port %s...", port)
	server := net.NewServer(":" + port)
	defer close(server.Msgch)
	defer close(server.Quit)
	defer net.CloseLogFile()
	server.Start()
}
