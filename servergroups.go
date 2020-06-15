package main

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
func (TSClient *Conn) ServerGroups() (*[]ServerGroup, error) {
	var serverGroups []ServerGroup

	// Get the server groups
	res, err := TSClient.Exec("servergrouplist")
	if err != nil || len(strings.Split(res, "\n")) <= 2 {
		return nil, err
	}

	// Parse server groups into []ServerGroups
	parts := strings.Split(strings.Split(res, "\n")[0], "|")
	for i := 0; i < len(parts); i++ {
		seg := strings.Split(parts[i], " ")

		id, err := strconv.ParseInt(GetVal(seg[0]), 10, 64)
		if err != nil {
			return nil, err
		}

		// Get the group ID so we can convert it into an Enum
		groupTypeId, err := strconv.ParseInt(GetVal(seg[2]), 10, 64)
		if err != nil {
			return nil, err
		}

		// Append the ServerGroup to the group array
		serverGroups = append(serverGroups, ServerGroup{
			Id:   id,
			Name: Decode(GetVal(seg[1])),
			Type: GroupType(groupTypeId),
		})
	}

	return &serverGroups, nil
}

// Add a client to a server group
func (TSClient *Conn) ServerGroupAddClient(sgid int, cldbid int) (*QueryResponse, error) {
	res, err := TSClient.Exec("servergroupaddclient sgid=%v cldbid=%v", sgid, cldbid)
	if err != nil {
		return nil, err
	}

	queryResponse := ParseQueryResponse(res)

	return &queryResponse, nil
}

// Remove a client from a server group
func (TSClient *Conn) ServerGroupRemoveClient(sgid int, cldbid int) (*QueryResponse, error) {
	res, err := TSClient.Exec("servergroupdelclient sgid=%v cldbid=%v", sgid, cldbid)
	if err != nil {
		return nil, err
	}

	queryResponse := ParseQueryResponse(res)
	return &queryResponse, nil
}

//List the users who belong to a specific server group
func (TSClient *Conn) ServerGroupMembers(gid int) (*[]User, error) {
	users := []User{}

	res, err := TSClient.Exec(fmt.Sprintf("servergroupclientlist sgid=%v", gid))
	if err != nil {
		return nil, err
	}

	// Map of active clients.
	// The key is their Database IDs (CLDBID)
	sessions, err := TSClient.ActiveClients()
	if err != nil {
		return nil, err
	}

	parts := strings.Split(strings.Split(res, "\n")[0], "|")
	for i := 0; i < len(parts); i++ {
		cldbid, err := strconv.ParseInt(strings.ReplaceAll(parts[i], "cldbid=", ""), 10, 64)
		if err != nil {
			return nil, err
		}

		// Get the user information using the
		// Client databse ID
		user, err := TSClient.FindUserByDbId(cldbid)
		if err != nil {
			return nil, err
		}

		// Assign the CLID (active client IDs) for the users active sessions
		user.ActiveSessionIds = sessions[user.Cldbid]

		// Add user object to result array
		users = append(users, *user)
	}

	return &users, nil
}

// Pokes all active clients belonging to databaseusers in a specific server group
func (TSClient *Conn) ServerGroupPoke(sgid int, msg string) error {
	// Encode the message once for all pokes
	msg = Encode(msg)
	// Get a list of users who belong to the specified GID (group)
	users, err := TSClient.ServerGroupMembers(gid)
	if err != nil {
		return err
	}

	for _, e := range *users {
		for ii := 0; ii < len(e.ActiveSessionIds); ii++ {
			TSClient.Exec("clientpoke clid=%v msg=%v", e.ActiveSessionIds[ii], msg)
		}
	}

	return nil
}
