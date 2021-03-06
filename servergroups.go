package ts3

import (
	"encoding/json"
	"fmt"
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
func ServerGroupMembers(sgid int64) (*status, []User, error) {
	queries := []KeyValue{
		{key: "sgid", value: i64tostr(sgid)},
	}

	qres, body, err := get("servergroupclientlist", false, queries)
	if err != nil || !qres.IsSuccess() {
		Log(Error, "Failed to get servergroup %v members \n%v\n%v", sgid, qres, err)
		return qres, nil, err
	}

	// For reasons that continue to baffle me
	// teamspeak devs use the term 'cldbid' in the servergroupclientlist
	// yet in other endpoints they use 'client_database_id'
	// so I need to have this struct for this one edge case....
	type cldbid_ struct {
		Clid int64 `json:"cldbid,string"`
	}

	var cldbid []cldbid_
	json.Unmarshal([]byte(body), &cldbid)

	qres1, sessions, err := ActiveClients()
	if err != nil {
		return qres1, nil, err
	}

	// Build an array of Users with their active session IDs (CLIDs) included
	groupmembers := []User{}
	for _, member := range cldbid {
		_, u, err := UserFindByDbId(member.Clid)
		if err != nil {
			Log(Error, "Failed to look up cldbid %v \n%v", member.Clid, err)
			continue
		}

		u.ActiveSessionIds = sessions[member.Clid]

		groupmembers = append(groupmembers, *u)
	}

	return qres, groupmembers, err
}

// // Pokes all active clients belonging to databaseusers in a specific server group
func ServerGroupPoke(sgid int64, msg string) (*status, error) {
	qres, users, err := ServerGroupMembers(sgid)
	if err != nil || !qres.IsSuccess() {
		Log(Error, "Failed to get server group members")
		return qres, err
	}

	attempted := 0
	failed := 0

	for _, user := range users {
		for i := 0; i < len(user.ActiveSessionIds); i++ {
			qres1, err := UserPoke(user.ActiveSessionIds[i], msg)
			if err != nil || !qres1.IsSuccess() {
				failed++
				Log(Error, "Failed to poke %v \n%v\n%v", user.Nickname, qres1, err)
			}

			attempted++
		}
	}

	qres.Code = -1
	qres.Message = fmt.Sprintf("%v%% of clients successfully poked (%v failed)", ((attempted-failed)/attempted)*100, failed)
	return qres, err
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
