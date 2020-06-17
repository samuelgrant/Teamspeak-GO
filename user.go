package ts3

import (
	"fmt"
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
func (TSClient *Conn) ActiveClients() (*QueryResponse, map[int64][]int64, error) {
	clients := make(map[int64][]int64)

	res, body, err := TSClient.Exec("clientlist")
	if err != nil || !res.IsSuccess {
		Log(Error, "Failed to get the active client list \n%v \n%v", res, err)
		return res, clients, err
	}

	//For each user
	users := strings.Split(body, "|")
	for i := 0; i < len(users); i++ {
		parts := strings.Split(users[i], " ")

		// Get the users client database id
		cldbid, err := strconv.ParseInt(GetVal(parts[2]), 10, 64)
		if err != nil {
			Log(Error, "Failed to parse the Client Database ID \n%v \n%v", parts, err)
			return res, clients, err
		}

		// Get the users client id (Active session ID)
		clid, err := strconv.ParseInt(GetVal(parts[0]), 10, 64)
		if err != nil {
			Log(Error, "Failed to parse the Client ID (active session ID) \n%v \n%v", parts, err)
			return res, clients, err
		}

		clients[cldbid] = append(clients[cldbid], clid)
	}

	return res, clients, nil
}

// Search for a user using the CLDBID and return a user object
func (TSClient *Conn) UserFindByDbId(cldbid int64) (*User, error) {
	res, body, err := TSClient.Exec("clientdbinfo cldbid=%v", cldbid)
	if err != nil || !res.IsSuccess {
		Log(Error, "Failed to get cldbid=%v information \n%v \n%v", cldbid, res, err)
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
func (TSClient *Conn) UserFindByCustomSearch(ident string, value string) (*QueryResponse, *User, error) {
	res, body, err := TSClient.Exec("customsearch ident=%v pattern=%v",
		// We must use underscores instead of spaces within the database
		// as the space is used to seperate sections of the telnet statment
		strings.ReplaceAll(ident, " ", "_"),
		strings.ReplaceAll(value, " ", "_"),
	)
	if err != nil || !res.IsSuccess {
		Log(Error, "Failed to find user with the custom field [%v]:%v \n%v \n%v", ident, value, res, err)
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
	user, err := TSClient.UserFindByDbId(cldbid)
	if err != nil {
		Log(Error, "Failed to look up the user %v \n%v", cldbid, err)
		return res, nil, nil
	}

	return res, user, nil
}

// Poke a client with a message
func (TSClient *Conn) UserPoke(clid int, msg string) (*QueryResponse, error) {
	res, _, err := TSClient.Exec("clientpoke clid=%v msg=%v", clid, Encode(msg))
	return res, err
}

// Delete a user from the user database. This will revoke all of their permissions
// and can be used to clear a users custom fields
func (TSClient *Conn) UserDelete(cldbid int) (*QueryResponse, error) {
	// Kick the users clients from the server
	res, err := TSClient.UserKickClients(cldbid, "Your access has been revoked; did you reset your Team Speak access?")
	if err != nil || !res.IsSuccess {
		Log(Error, "Failed to kick all clients belonging to the user %v \n%v\n%v", cldbid, res, err)
		return res, err
	}

	// Delete the users account
	res, _, err = TSClient.Exec("clientdbdelete cldbid=%v", cldbid)
	if err != nil || !res.IsSuccess {
		Log(Error, "Failed to delete user from the TS database \n%v\n%v", res, err)
		return res, err
	}

	return res, nil
}

// Kick a users clients from the server
func (TSClient *Conn) UserKickClients(cldbid int, msg string) (*QueryResponse, error) {
	res, sessions, err := TSClient.ActiveClients()
	if err != nil || !res.IsSuccess {
		Log(Error, "Unable to get active clients from the server \n%v\n%v", res, err)
		return res, err
	}

	msg = Encode(msg)
	var failures int = 0
	var counter int = 0

	// Kick all active sessions belonging to the user (client database id)
	for _, clid := range sessions[int64(cldbid)] {
		res, _, err := TSClient.Exec("clientkick clid=%v reasonid=%v reasonmsg=%v", clid, 5, msg)
		if err != nil || !res.IsSuccess {
			Log(Error, "Unable to kick %v client session id (CLID) \n%v\n%v", cldbid, clid, res, err)
			failures++
		}

		counter++
	}

	// Return a custom Query Response
	// -1 custom response ID (teamspeak uses positive values)
	return &QueryResponse{
		Id:        -1,
		Msg:       fmt.Sprintf("Sucesfully kicked %v out of %v clients", (counter - failures), counter),
		IsSuccess: 0 == failures,
	}, nil
}
