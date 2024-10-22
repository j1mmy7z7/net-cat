package main

import (
	"log"
	"os"
)

var logFile *os.File

// Open a new log file for appending
func init(){
	var err error
	logFile,err = os.OpenFile("chat.log",os.O_APPEND|os.O_CREATE|os.O_WRONLY,0666)
	if err != nil {
log.Fatal(err)
	}
}

// Log messages to the file
func logMessage(msg string){
	if logFile!=nil{
		_,err:=logFile.WriteString(msg+"\n")
		if err!=nil{
			log.Println("Error writing to log",err)
		}
	}
}
// close the log file before exiting the program
func closeLogFile(){
	if logFile!=nil{
		logFile.Close()
	}
}