package ts3

import (
	"encoding/json"
	"fmt"
	"strings"
)

type User struct {
	// Database ID
	Cldbid int64 `json:"client_database_id,string"`
	// Unique TS ID
	Cluid string `json:"client_unique_identifier"`
	// Last Connection
	LastConnected string `json:"client_lastconnected"`
	LastIP        string `json:"client_lastip"`

	ActiveSessionIds []int64
	Nickname         string `json:"client_nickname"`
}

type Session struct {
	Cldbid   int64  `json:"client_database_id,string"`
	Clid     int64  `json:"clid,string"`
	Nickname string `json:"client_nickname"`
}

// Returns a map of active sessions mapped to their database IDs
func ActiveClients() (*status, map[int64][]int64, error) {
	qres, body, err := get("clientlist", false)
	if err != nil || !qres.IsSuccess() {
		Log(Error, "Failed to get the active client list \n%v\n%v", qres, err)
		return qres, nil, err
	}

	var sessions []Session
	json.Unmarshal([]byte(body), &sessions)

	// Map of sessions by Client DB ID
	connections := make(map[int64][]int64)
	for _, session := range sessions {
		connections[session.Cldbid] = append(connections[session.Cldbid], session.Clid)
	}

	return qres, connections, err
}

// Search for a user using the CLDBID and return a user object
func UserFindByDbId(cldbid int64) (*status, *User, error) {
	queries := []KeyValue{
		{key: "cldbid", value: i64tostr(cldbid)},
	}

	qres, body, err := get("clientdbinfo", false, queries)
	if err != nil || !qres.IsSuccess() {
		Log(Error, "Failed to get information for CLDBID %v \n%v\n%v", cldbid, qres, err)
		return qres, nil, err
	}

	var user []User
	json.Unmarshal([]byte(body), &user)
	return qres, &user[0], err
}

// Find a user using the custom field sets that were attached to their privilege token
// You can only search one column/ident and value at a time.
func UserFindByCustomSearch(ident string, pattern string) (*status, *User, error) {
	queries := []KeyValue{
		{key: "ident", value: strings.ReplaceAll(ident, " ", "_")},
		{key: "pattern", value: strings.ReplaceAll(pattern, " ", "_")},
	}

	qres, body, err := get("customsearch", false, queries)
	if err != nil || !qres.IsSuccess() {
		Log(Error, "Failed to find user using params {ident: %v, pattern: %v} \n%v\n%v", ident, pattern, qres, err)
		return qres, nil, err
	}

	// For reasons that continue to baffle me
	// teamspeak devs use the term 'cldbid' in the servergroupclientlist
	// yet in other endpoints they use 'client_database_id'
	// so I need to have this struct for this one edge case....
	type cldbid_ struct {
		Clid int64 `json:"cldbid,string"`
	}

	var u []cldbid_
	json.Unmarshal([]byte(body), &u)
	qres, user, err := UserFindByDbId(u[0].Clid)
	if err != nil || !qres.IsSuccess() {
		Log(Error, "Failed to get user information for CLDBID %v \n%v\n%v", u[0].Clid, qres, err)
		return qres, nil, err
	}

	// Get the sssion information
	qres, sessions, err := ActiveClients()
	if err != nil || !qres.IsSuccess() {
		Log(Error, "Failed to get the active sessions list \n%v\n%v", qres, err)
	}

	// Attach the users sessions
	user.ActiveSessionIds = sessions[user.Cldbid]
	return qres, user, err
}

// Poke a client with a message
func UserPoke(clid int64, msg string) (*status, error) {
	queries := []KeyValue{
		{key: "clid", value: i64tostr(clid)},
		{key: "msg", value: msg},
	}

	qres, _, err := get("clientpoke", false, queries)
	if err != nil || !qres.IsSuccess() {
		Log(Error, "Failed to poke CLID %v \n%v\n%v", clid, qres, err)
	}

	return qres, err
}

// Delete a user from the user database. This will revoke all of their permissions
// and can be used to clear a users custom fields
func UserDelete(cldbid int64) (*status, error) {
	// We need to kick their clients from the server before we can delete their account
	qres, err := UserKickClients(cldbid, "Your access has been revoked; did you reset your Team Speak access?")
	if err != nil || !qres.IsSuccess() {
		Log(Error, "Failed to kick all clients belonging to user (CLDBID %v) \n%v\n%v", cldbid, qres, err)
		return qres, err
	}

	queries := []KeyValue{
		{key: "cldbid", value: i64tostr(cldbid)},
	}
	qres, _, err = get("clientdbdelete", false, queries)
	if err != nil || !qres.IsSuccess() {
		Log(Error, "Failed to delete user from the TeamSpeak database \n%v\n%v", qres, err)
	}

	return qres, err
}

// Kick a users clients from the server
func UserKickClients(cldbid int64, msg string) (*status, error) {
	qres, sessions, err := ActiveClients()
	if err != nil || !qres.IsSuccess() {
		Log(Error, "Failed to get the active sessions \n%v\n%v", qres, err)
		return qres, err
	}

	attempted := 0
	failed := 0

	for _, clid := range sessions[cldbid] {
		queries := []KeyValue{
			{key: "clid", value: i64tostr(clid)},
			{key: "reasonid", value: "5"},
			{key: "reasonmsg", value: msg},
		}

		qres1, _, err := get("clientkick", false, queries)
		if err != nil || !qres.IsSuccess() {
			failed++
			Log(Error, "Failed to kick CLID %v \n%v\n%v", clid, qres1, err)
		}

		attempted++
	}

	qres.Code = -1
	qres.Message = fmt.Sprintf("%v%% of clients successfully kicked from the server (%v failed)", ((attempted-failed)/attempted)*100, failed)
	return qres, err
}
