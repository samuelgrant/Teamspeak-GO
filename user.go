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

// Returns a map of active sessions mapped to their database IDs
func (TSClient *Conn) ActiveClients() (map[int64][]int64, error) {
	clients := make(map[int64][]int64)

	res, err := TSClient.Exec("clientlist")
	if err != nil || len(strings.Split(res, "\n")) <= 2 {
		return clients, err
	}

	parts := strings.Split(res, "|")
	for i := 0; i < len(parts); i++ {
		c := strings.Split(parts[i], " ")

		cldbid, err := strconv.ParseInt(GetVal(c[2]), 10, 64)
		if err != nil {
			return clients, err
		}

		clid, err := strconv.ParseInt(GetVal(c[0]), 10, 64)
		if err != nil {
			return clients, err
		}

		clients[cldbid] = append(clients[cldbid], clid)
	}

	return clients, nil
}

// Search for a user using the CLDBID and return a user object
func (TSClient *Conn) FindUserByDbId(cldbid int64) (*User, error) {
	res, err := TSClient.Exec("clientdbinfo cldbid=%v", cldbid)
	if err != nil || len(strings.Split(res, "\n")) <= 2 {
		return nil, err
	}

	// Return the user object including:
	// Database ID, Unique Client ID, Last Connected (time/ip), Nickname
	parts := strings.Split(strings.Split(res, "\n")[0], " ")
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
	res, err := TSClient.Exec("customsearch ident=%v pattern=%v",
		strings.ReplaceAll(ident, " ", "_"),
		strings.ReplaceAll(value, " ", "_"),
	)
	if err != nil {
		return nil, nil, err
	}

	lines := strings.Split(res, "\n")
	queryResponse := ParseQueryResponse(lines[len(lines)-2])
	if len(lines) <= 2 {
		return &queryResponse, &User{}, err
	}

	cldbid, err := strconv.ParseInt(strings.Split(GetVal(lines[0]), " ")[0], 10, 64)
	if err != nil {
		return &queryResponse, &User{}, err
	}

	user, err := TSClient.FindUserByDbId(cldbid)
	return &queryResponse, user, nil
}
