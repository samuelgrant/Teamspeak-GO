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
func (TSClient *Conn) ServerGroupAddClient(sgid int, cldbid int) (*QueryResponse, error) {
	res, _, err := TSClient.Exec("servergroupaddclient sgid=%v cldbid=%v", sgid, cldbid)
	if err != nil || !res.IsSuccess {
		Log(Error, "Failed to add user %v to server group %v \n%v \n%v", cldbid, sgid, res, err)
		return res, err
	}

	return res, nil
}

// Remove a client from a server group
func (TSClient *Conn) ServerGroupRemoveClient(sgid int, cldbid int) (*QueryResponse, error) {
	res, _, err := TSClient.Exec("servergroupdelclient sgid=%v cldbid=%v", sgid, cldbid)
	if err != nil || !res.IsSuccess {
		Log(Error, "Failed to remove user %v from server group %v \n%v \n%v", cldbid, sgid, res, err)
		return res, err
	}

	return res, nil
}

//List the users who belong to a specific server group
func (TSClient *Conn) ServerGroupMembers(gid int) (*QueryResponse, *[]User, error) {
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
func (TSClient *Conn) ServerGroupPoke(sgid int, msg string) (*QueryResponse, error) {
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
			res, err := TSClient.UserPoke(int(user.ActiveSessionIds[i]), msg)
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
