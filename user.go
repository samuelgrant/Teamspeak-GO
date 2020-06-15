package main

import (
	"strconv"
	"strings"
)

type User struct {
	// Database ID
	Cldbid int64
	// Unique TS ID
	Cluid string
	// Last Connection
	LastConnected string
	LastIP        string

	ActiveSessionIds []int64
	Nickname         string
}

// func (TSClient *Conn) FindByName(nickname string) {}

// Returns a map of active sessions mapped to their database IDs
func (TSClient *Conn) ActiveClients() (map[int64][]int64, error) {
	clients := make(map[int64][]int64)

	res, body, err := TSClient.Exec("clientlist")
	if err != nil || !res.IsSuccess {
		Log(Error, "Failed to get the active client list")
		return clients, err
	}

	//For each user
	users := strings.Split(body, "|")
	for i := 0; i < len(users); i++ {
		parts := strings.Split(users[i], " ")

		// Get the users client database id
		cldbid, err := strconv.ParseInt(GetVal(parts[2]), 10, 64)
		if err != nil {
			Log(Error, "Failed to parse the Client Database ID\n%v\n", parts, err)
			return clients, err
		}

		// Get the users client id (Active session ID)
		clid, err := strconv.ParseInt(GetVal(parts[0]), 10, 64)
		if err != nil {
			Log(Error, "Failed to parse the Client ID (active session ID)\n%v\n", parts, err)
			return clients, err
		}

		clients[cldbid] = append(clients[cldbid], clid)
	}

	return clients, nil
}

// Search for a user using the CLDBID and return a user object
func (TSClient *Conn) FindUserByDbId(cldbid int64) (*User, error) {
	res, body, err := TSClient.Exec("clientdbinfo cldbid=%v", cldbid)
	if err != nil || !res.IsSuccess {
		Log(Error, "Failed to get cldbid=%v information\n%v\n%v", cldbid, res, err)
		return nil, err
	}

	// Return the user object including:
	// Database ID, Unique Client ID, Last Connected (time/ip), Nickname
	parts := strings.Split(body, " ")
	return &User{
		Cldbid:        cldbid,
		Cluid:         GetVal(parts[0]),
		LastConnected: GetVal(parts[4]),
		LastIP:        GetVal(parts[13]),
		Nickname:      Decode(GetVal(parts[1])),
	}, nil
}

// Find a user using the custom field sets that were attached to their privilege token
// You can only search one column/ident and value at a time.
func (TSClient *Conn) FindUserByCustomSearch(ident string, value string) (*QueryResponse, *User, error) {
	res, body, err := TSClient.Exec("customsearch ident=%v pattern=%v",
		// We must use underscores instead of spaces within the database
		// as the space is used to seperate sections of the telnet statment
		strings.ReplaceAll(ident, " ", "_"),
		strings.ReplaceAll(value, " ", "_"),
	)
	if err != nil || !res.IsSuccess {
		Log(Error, "Failed to find user with the custom field [%v]:%v", ident, value)
		return res, nil, err
	}

	parts := strings.Split(body, " ")

	// Get the client database ID
	cldbid, err := strconv.ParseInt(GetVal(parts[0]), 10, 64)
	if err != nil {
		Log(Error, "Failed to parse the CLDBID from the query response body \n%v\n%v", res, err)
		return res, nil, err
	}

	// Look up the user using their client database ID
	user, err := TSClient.FindUserByDbId(cldbid)
	if err != nil {
		Log(Error, "Failed to look up the user %v", cldbid)
		return res, nil, nil
	}

	return res, user, nil
}
