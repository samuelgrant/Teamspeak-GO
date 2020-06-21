package ts3

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

type GroupType int

const (
	TemplateGroup GroupType = 0
	RegularGroup  GroupType = 1
	QueryGroup    GroupType = 2
)

type ServerGroup struct {
	Id   int64     `json:"sgid,string"`
	Name string    `json:"name"`
	Type GroupType `json:"type,string"`
}

// List all of the server groups on the server
func ServerGroups() (*status, []ServerGroup, error) {
	qres, body, err := get("servergrouplist", false)
	if err != nil || !qres.IsSuccess() {
		Log(Error, "Failed to get a list server groups")
		return qres, nil, err
	}

	var groups []ServerGroup
	json.Unmarshal([]byte(body), &groups)

	return qres, groups, err
}

// Add a client to a server group
func ServerGroupsAddClient(sgid int64, cldbid int64) (*status, error) {
	queries := []KeyValue{
		{key: "sgid", value: i64tostr(sgid)},
		{key: "cldbid", value: i64tostr(cldbid)},
	}

	qres, _, err := get("servergroupaddclient", false, queries)
	if err != nil || !qres.IsSuccess() {
		Log(Error, "Failed to assign servergroup %v to clientdbid %v \n%v\n%v", sgid, cldbid, qres, err)
	}

	return qres, err
}

// Remove a client from a server group
func ServerGroupsRevokeClient(sgid int64, cldbid int64) (*status, error) {
	queries := []KeyValue{
		{key: "sgid", value: i64tostr(sgid)},
		{key: "cldbid", value: i64tostr(cldbid)},
	}

	qres, _, err := get("servergroupdelclient", false, queries)
	if err != nil || !qres.IsSuccess() {
		Log(Error, "Failed to revoke servergroup %v from cldboid %v \n%v\n%v", sgid, cldbid, qres, err)
	}

	return qres, err
}

//List the users who belong to a specific server group
func (TSClient *Conn) ServerGroupMembers(gid int64) (*QueryResponse, *[]User, error) {
	users := []User{}

	// Get the list of server group clients from TS
	// This list returns a CLDBID, this isn't of huge use so we need to look up more info
	res, body, err := TSClient.Exec(fmt.Sprintf("servergroupclientlist sgid=%v", gid))
	if err != nil || !res.IsSuccess {
		Log(Error, "Failed to get the server group client list \n%v \n%v", res, err)
		return res, nil, err
	}

	// Map of active clients using their DatabseID as the map key
	res, sessions, err := TSClient.ActiveClients()
	if err != nil || !res.IsSuccess {
		Log(Error, "Failed to get the list of active clients \n%v \n%v", res, err)
		return res, nil, err
	}

	parts := strings.Split(body, "|")
	for i := 0; i < len(parts); i++ {
		cldbid, err := strconv.ParseInt(strings.ReplaceAll(parts[i], "cldbid=", ""), 10, 64)
		if err != nil {
			Log(Error, "Error parsing the CLDBID \n%v \n%v", res, err)
			return res, nil, err
		}

		// Get the user information using their Client DB ID
		user, err := TSClient.UserFindByDbId(cldbid)
		if err != nil {
			Log(Error, "Error finding the user by their cldbid \n%v \n%v", res, err)
			return res, nil, err
		}

		// Assign the CLID (active client IDs) for the users active sessions
		user.ActiveSessionIds = sessions[user.Cldbid]

		// Add user object to result array
		users = append(users, *user)
	}

	return res, &users, nil
}

// Pokes all active clients belonging to databaseusers in a specific server group
func (TSClient *Conn) ServerGroupPoke(sgid int64, msg string) (*QueryResponse, error) {
	// Get a list of users who belong to the specified GID (group)
	res, body, err := TSClient.ServerGroupMembers(sgid)
	if err != nil || !res.IsSuccess {
		Log(Error, "Error getting a list of server group members \n%v \n%v", res, err)
		return res, err
	}

	var successful int = 0
	var attempted int = 0

	for _, user := range *body {
		for i := 0; i < len(user.ActiveSessionIds); i++ {
			res, err := TSClient.UserPoke(user.ActiveSessionIds[i], msg)
			if err != nil {
				Log(Error, "Failed to poke %v \n%v \n%v", user.Nickname, res, err)
			}

			// Increase counters
			attempted++
			if res.IsSuccess {
				successful++
			}
		}
	}

	return &QueryResponse{
		Id:        -1,
		Msg:       fmt.Sprintf("%v out of %v clients sucesffuly poked", successful, attempted),
		IsSuccess: true,
	}, nil
}

// List a users server groups
func ServerGroupsByClientDbId(cldbid int64) (*status, []ServerGroup, error) {
	queries := []KeyValue{
		{key: "cldbid", value: i64tostr(cldbid)},
	}

	qres, body, err := get("servergroupsbyclientid", false, queries)
	if err != nil || !qres.IsSuccess() {
		Log(Error, "Failed to get a list of cldbid %v servergroups \n%v\n%v", cldbid, qres, err)
		return qres, nil, err
	}

	var groups []ServerGroup
	json.Unmarshal([]byte(body), &groups)
	return qres, groups, err
}

// Creates a server group
func ServerGroupAdd(name string) (*status, int64, error) {
	queries := []KeyValue{
		{key: "name", value: name},
	}

	qres, body, err := get("servergroupadd", false, queries)
	if err != nil || !qres.IsSuccess() {
		Log(Error, "Failed to create a new servergroup \n%v\n%v", qres, err)
		return qres, -1, err
	}

	var group []ServerGroup
	json.Unmarshal([]byte(body), &group)
	return qres, group[0].Id, err
}

// Create a duplicate of the server group {ssgid}. The new group will be named {name}
func ServerGroupCopy(ssgid int64, name string) (*status, int64, error) {
	queries := []KeyValue{
		{key: "ssgid", value: i64tostr(ssgid)},
		{key: "tsgid", value: "0"}, // We want to make a new group
		{key: "name", value: name},
		{key: "type", value: i64tostr(int64(RegularGroup))},
	}

	qres, body, err := get("servergroupcopy", false, queries)
	if err != nil || !qres.IsSuccess() {
		Log(Error, "Failed to copy group %v \n%v\n%v", ssgid, qres, err)
		return qres, -1, err
	}

	var group []ServerGroup
	json.Unmarshal([]byte(body), &group)
	return qres, group[0].Id, err
}

// Delete a server group, forceDelete deletes a group with members
func ServerGroupDel(sgid int64, forceDelete bool) (*status, error) {
	var force int64 = 0
	if forceDelete {
		force = 1
	}

	queries := []KeyValue{
		{key: "sgid", value: i64tostr(sgid)},
		{key: "force", value: i64tostr(force)},
	}

	qres, _, err := get("servergroupdel", false, queries)
	if err != nil || !qres.IsSuccess() {
		Log(Error, "Failed to delete servergroup %v \n%v\n%v", sgid, qres, err)
	}

	return qres, err
}
