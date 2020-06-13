package main

import (
	"log"
)

const (
	InfoColor    = "\033[1;34m%s\033[0m"
	NoticeColor  = "\033[1;36m%s\033[0m"
	WarningColor = "\033[1;33m%s\033[0m"
	ErrorColor   = "\033[1;31m%s\033[0m"
	DebugColor   = "\033[0;36m%s\033[0m"
)

/**
* This file is being used as a way of testing code during development.
* DONOT Commit credentials which are in line 21
*
* TODO: Clean this file when a better solution for testing is developed.
 */
func main() {
	var server, user, passwd = "", "", ""
	log.Printf(NoticeColor, "[info]: Starting worker")

	TSClient, err := Connect(server)
	if err != nil {
		log.Fatalf("[error]: Failed to connect to the TS server: %v", err)
	}

	err = TSClient.Login(user, passwd)
	if err != nil {
		log.Fatal(ErrorColor, "[error]: Failed to authenticate with the server")
	}

	_, err = TSClient.Use(1)

	// To Document
	TSClient.Disconnect()
	_ = TSClient.IsConnected()

	for {
	}
}
