package ts3

import (
	"encoding/json"
	"fmt"
)

type ChannelGroup struct {
	Id   int64     `json:"cgid,string"`
	Name string    `json:"name"`
	Type GroupType `json:"type,string"` // found in servergroup.go
}

// Get a list of channel groups on the server
func ChannelGroups() (*status, []ChannelGroup, error) {
	qres, body, err := get("channelgrouplist", false)
	if err != nil || !qres.IsSuccess() {
		Log(Error, "Failed to get a list of channelgroups \n%v\n%v", qres, err)
		return qres, nil, err
	}

	var groups []ChannelGroup
	json.Unmarshal([]byte(body), &groups)
	return qres, groups, err
}

// Add a client to a specific channel group for a given channel
func SetChannelGroup(cgid int64, cid int64, cldbid int64) (*status, error) {
	queries := []KeyValue{
		{key: "cgid", value: i64tostr(cgid)},
		{key: "cid", value: i64tostr(cid)},
		{key: "cldbid", value: i64tostr(cldbid)},
	}

	qres, _, err := get("setclientchannelgroup", false, queries)
	if err != nil || !qres.IsSuccess() {
		Log(Error, "Unable to update client channel group {cldbid: %v, cid: %v, cgid: %v} \n%v\n%v", cldbid, cid, cgid, qres, err)
	}

	return qres, err
}

// Set a user back to the default channel group. The default channel group is a group that:
// a) has the name "guest" && b) is a RegularGroup type
func ResetChannelGroup(cid int64, cldbid int64) (*status, error) {
	qres, groups, err := ChannelGroups()
	if err != nil || !qres.IsSuccess() {
		Log(Error, "Failed to get avaliable channel groups \n%v\n%v", qres, err)
		return qres, err
	}

	var cgid int64
	for _, group := range groups {
		if group.Name == "Guest" && group.Type == RegularGroup {
			cgid = group.Id
			break
		}
	}

	return SetChannelGroup(cgid, cid, cldbid)
}

// Return the members of a specific channel group for a given channel
func ChannelGroupMembers(cgid int64, cid int64) (*status, []User, error) {
	queries := []KeyValue{
		{key: "cid", value: i64tostr(cid)},
		{key: "cgid", value: i64tostr(cgid)},
	}

	qres, body, err := get("channelgroupclientlist", false, queries)
	if err != nil || !qres.IsSuccess() {
		Log(Error, "Failed to get channelgroup members \n%v\n%v", qres, err)
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

// Poke all clients who belong to a given channel group in a specific channel
func ChannelGroupPoke(cgid int64, cid int64, msg string) (*status, error) {
	qres, members, err := ChannelGroupMembers(cgid, cid)
	if err != nil || !qres.IsSuccess() {
		Log(Error, "Failed to get group members")
		return qres, err
	}

	attempted := 0
	failed := 0

	for _, user := range members {
		for i := 0; i < len(user.ActiveSessionIds); i++ {
			res, err := UserPoke(user.ActiveSessionIds[i], msg)
			if err != nil {
				failed++
				Log(Error, "Failed to poke %v \n%v\n%v", user.Nickname, res, qres)
			}

			attempted++
		}
	}

	qres.Code = -1
	qres.Message = fmt.Sprintf("%v%% of clients successfully poked (%v failed)", ((attempted-failed)/attempted)*100, failed)
	return qres, err
}
