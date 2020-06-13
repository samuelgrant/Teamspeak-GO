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

	err = TSClient.Authenticate(user, passwd)
	if err != nil {
		log.Fatal(ErrorColor, "[error]: Failed to authenticate with the server")
	}

	_, err = TSClient.Use(1)

	// res, token, err := TSClient.TokensAdd(7, "This is a test description", map[string]string{"auth id": "37", "auth name": "samuel the horiable"})
	// if err != nil {
	// 	log.Fatalf("Failed to create token: %v", err)
	// }
	// log.Printf("Res: %v\nToken: %v", res, token)

	// Delete token: ttO+GWv3bYcXJJAnrd7Jb0b5e+irv7AxsSRAeg69
	// res, err = TSClient.TokensDelete("ttO+GWv3bYcXJJAnrd7Jb0b5e+irv7AxsSRAeg69")
	// if err != nil {
	// 	log.Fatalf("Failed to delete token: %v", err)
	// }
	// log.Printf("Res: %v", res)

	// List tokens
	_, tokens, err := TSClient.Tokenslist()
	if err != nil {
		log.Fatalf("Failed to list tokens")
	}
	log.Fatal(tokens)

	// To Document
	TSClient.Disconnect()
	_ = TSClient.IsConnected()

	for {
	}
}
