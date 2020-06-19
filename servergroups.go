package ts3

import (
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
	Id   int64
	Name string
	Type GroupType
	// IconId   int
	// SaveDb   int
	// SortId   int
	// NameMode string
}

// List all of the server groups on the server
func (TSClient *Conn) ServerGroups() (*QueryResponse, *[]ServerGroup, error) {
	var serverGroups []ServerGroup

	// Get the server groups
	res, body, err := TSClient.Exec("servergrouplist")
	if err != nil || !res.IsSuccess {
		return res, nil, err
	}

	// Parse server groups into []ServerGroups
	parts := strings.Split(body, "|")
	for i := 0; i < len(parts); i++ {
		seg := strings.Split(parts[i], " ")

		id, err := strconv.ParseInt(GetVal(seg[0]), 10, 64)
		if err != nil {
			Log(Error, "Failed to parse the server group ID \n%v", err)
			return res, nil, err
		}

		// Get the group ID so we can convert it into an Enum
		groupTypeId, err := strconv.ParseInt(GetVal(seg[2]), 10, 64)
		if err != nil {
			Log(Error, "Failed to parse the group type ID \n%v", err)
			return res, nil, err
		}

		// Append the ServerGroup to the group array
		serverGroups = append(serverGroups, ServerGroup{
			Id:   id,
			Name: Decode(GetVal(seg[1])),
			Type: GroupType(groupTypeId),
		})
	}

	return res, &serverGroups, nil
}

// Add a client to a server group
func (TSClient *Conn) ServerGroupAddClient(sgid int64, cldbid int64) (*QueryResponse, error) {
	res, _, err := TSClient.Exec("servergroupaddclient sgid=%v cldbid=%v", sgid, cldbid)
	if err != nil || !res.IsSuccess {
		Log(Error, "Failed to add user %v to server group %v \n%v \n%v", cldbid, sgid, res, err)
		return res, err
	}

	return res, nil
}

// Remove a client from a server group
func (TSClient *Conn) ServerGroupRemoveClient(sgid int64, cldbid int64) (*QueryResponse, error) {
	res, _, err := TSClient.Exec("servergroupdelclient sgid=%v cldbid=%v", sgid, cldbid)
	if err != nil || !res.IsSuccess {
		Log(Error, "Failed to remove user %v from server group %v \n%v \n%v", cldbid, sgid, res, err)
		return res, err
	}

	return res, nil
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
func (TSClient *Conn) ServerGroupsByClientDbId(cldbid int64) (*QueryResponse, *[]ServerGroup, error) {
	var groups []ServerGroup

	res, body, err := TSClient.Exec("servergroupsbyclientid cldbid=%v", cldbid)
	if err != nil || !res.IsSuccess {
		Log(Error, "Failed to get %v's server groups \n%v\n%v", cldbid, res, err)
		return res, nil, err
	}

	lines := strings.Split(body, "|")
	for _, line := range lines {
		parts := strings.Split(line, " ")

		sgid, err := strconv.ParseInt(GetVal(parts[1]), 10, 64)
		if err != nil {
			Log(Error, "Failed to parse the server group id \n%v\n%v", res, err)
			return res, nil, err
		}

		groups = append(groups, ServerGroup{
			Id:   sgid,
			Name: Decode(GetVal(parts[0])),
			Type: 1,
		})
	}

	return res, &groups, nil
}

// Creates a server group
func (TSClient *Conn) ServerGroupAdd(name string) (*QueryResponse, int64, error) {
	res, body, err := TSClient.Exec("servergroupadd name=%v", Encode(name))
	if err != nil || !res.IsSuccess {
		Log(Error, "Failed to create a server group with the name %v \n%v\n%v", name, res, err)
		return res, -1, err
	}

	sgid, err := strconv.ParseInt(GetVal(body), 10, 64)
	if err != nil {
		Log(Error, "Failed to parse a server group ID")
		return res, -1, err
	}

	return res, sgid, nil
}

// Create a duplicate of the server group {ssgid}. The new group will be named {newName}
func (TSClient *Conn) ServerGroupCopy(ssgid int64, newName string) (*QueryResponse, int64, error) {
	// Duplicate the server group
	res, body, err := TSClient.Exec("servergroupcopy ssgid=%v tsgid=0 name=%v type=%v", ssgid, Encode(newName), RegularGroup)
	if err != nil || !res.IsSuccess {
		Log(Error, "Failed to copy servergroup %v \n%v\n%v", ssgid, res, err)
		return res, -1, err
	}

	// Get the new server group ID
	newSGID, err := strconv.ParseInt(GetVal(body), 10, 64)
	if err != nil {
		Log(Error, "Failed to parse the new server group ID")
		return res, -1, err
	}

	return res, newSGID, nil
}

// Delete a server group, set force delete to true to delete a group.
// with members
func (TSClient *Conn) ServerGroupDel(sgid int64, forceDelete bool) (*QueryResponse, error) {
	var force int64
	switch forceDelete {
	case true:
		force = 1
	case false:
		force = 0
	}

	// Attempt to delete the server group
	res, _, err := TSClient.Exec("servergroupdel sgid=%v force=%v", sgid, force)
	if err != nil || !res.IsSuccess {
		Log(Error, "Failed to delete the server group %v sgod \n%v\n%v", sgid, res, err)
		return res, err
	}

	return res, nil
}
