package ts3

type ChannelGroup struct {
	Id   int64
	Name string
	Type GroupType // found in servergroup.go
	// IconId   int
	// SaveDb   int
	// SortId   int
	// NameMode int
}

// Get a list of channel groups on the server
// func (TSClient *Conn) ChannelGroups() (*QueryResponse, *[]ChannelGroup, error) {
// 	var channelGroups []ChannelGroup

// 	res, body, err := TSClient.Exec("channelgrouplist")
// 	if err != nil || !res.IsSuccess {
// 		Log(Error, "Failed to get channel groups from TS server \n%v \n%v", res, err)
// 		return res, nil, err
// 	}

// 	groups := strings.Split(body, "|")
// 	for i := 0; i < len(groups); i++ {
// 		seg := strings.Split(groups[i], " ")

// 		// Get the channel group ID
// 		cgid, err := strconv.ParseInt(GetVal(seg[0]), 10, 64)
// 		if err != nil {
// 			Log(Error, "Failed to parse the Group ID \n%v", err)
// 			return nil, nil, err
// 		}

// 		// Get the Group Type ID so we can convert it to an enum later
// 		groupTypeId, err := strconv.ParseInt(GetVal(seg[2]), 10, 64)
// 		if err != nil {
// 			Log(Error, "Failed to parse the Group Type \n%v", err)
// 		}

// 		channelGroups = append(channelGroups, ChannelGroup{
// 			Id:   cgid,
// 			Type: GroupType(groupTypeId),
// 			Name: Decode(GetVal(seg[1])),
// 		})
// 	}

// 	return res, &channelGroups, nil
// }

// Add a client to a specific channel group for a given channel
// func (TSClient *Conn) SetChannelGroup(cgid int64, cid int64, cldbid int64) (*QueryResponse, error) {
// 	res, _, err := TSClient.Exec("setclientchannelgroup cgid=%v cid=%v cldbid=%v", cgid, cid, cldbid)
// 	if err != nil || !res.IsSuccess {
// 		Log(Error, "Unable to update client channel group {clientDbId: %v, channelId: %v, channelGroupId: %v}. \n%v \n%v", cldbid, cid, cgid, res, err)
// 		return res, err
// 	}

// 	return res, nil
// }

// Set a user back to the default channel group. The default channel group is a group that:
// a) has the name "guest" && b) is a RegularGroup type
// func (TSClient *Conn) ResetChannelGroup(cid int64, cldbid int64) (*QueryResponse, error) {
// 	res, groups, err := TSClient.ChannelGroups()
// 	if err != nil || !res.IsSuccess {
// 		Log(Error, "Failed to get avaliable channel groups \n%v \n%v", res, err)
// 		return res, err
// 	}

// 	// Find the group ID of the guest channel group
// 	var cgid int64
// 	for _, group := range *groups {
// 		if group.Name == "Guest" && group.Type == RegularGroup {
// 			cgid = group.Id
// 		}
// 	}

// 	// Set the channel group
// 	res, err = TSClient.SetChannelGroup(cgid, cid, cldbid)
// 	if err != nil || !res.IsSuccess {
// 		Log(Error, "Failed to reset the users channel group \n%v \n%v", res, err)
// 		return res, err
// 	}

// 	return res, nil
// }

// Return the members of a specific channel group for a given channel
// func (TSClient *Conn) ChannelGroupMembers(cgid int64, cid int64) (*QueryResponse, []User, error) {
// 	// users := []User{}

// 	// // Get a list of channel group members for a given channel
// 	// res, body, err := TSClient.Exec("channelgroupclientlist cid=%v cgid=%v", cid, cgid)
// 	// if err != nil || !res.IsSuccess {
// 	// 	Log(Error, "Failed to get the list of channel group members {cid: %v, cgid: %v} \n%v \n%v", cid, cgid, res, err)
// 	// 	return res, nil, err
// 	// }

// 	// // Map of active clients using their DatabaseID as the map key
// 	// // res, sessions, err := TSClient.ActiveClients()
// 	// if err != nil || !res.IsSuccess {
// 	// 	Log(Error, "Failed to get the list of active clients \n%v \n%v", res, err)
// 	// 	return res, nil, err
// 	// }

// 	// clients := strings.Split(body, "|")
// 	// for i := 0; i < len(clients); i++ {
// 	// 	parts := strings.Split(clients[i], " ")

// 	// 	// Parse the client database id
// 	// 	cldbid, err := strconv.ParseInt(GetVal(parts[1]), 10, 64)
// 	// 	if err != nil {
// 	// 		Log(Error, "Failed to parse the client database ID \n%v", err)
// 	// 		return nil, nil, err
// 	// 	}

// 	// 	// Find the user using their CLDBID
// 	// 	user, err := TSClient.UserFindByDbId(cldbid)
// 	// 	if err != nil {
// 	// 		Log(Error, "Failed to find user using the cldbid %v \n%v", cldbid, err)
// 	// 		return nil, nil, err
// 	// 	}

// 	// 	// Assign the CLID (active client IDs) for the users active sessions
// 	// 	// user.ActiveSessionIds = sessions[user.Cldbid]

// 	// 	// Add user object to result array
// 	// 	users = append(users, *user)
// 	// }

// 	return nil, nil, nil
// }

// Poke all clients who belong to a given channel group in a specific channel
// func (TSClient *Conn) ChannelGroupPoke(cgid int64, cid int64, msg string) (*QueryResponse, error) {
// 	res, body, err := TSClient.ChannelGroupMembers(cgid, cid)
// 	if err != nil || !res.IsSuccess {
// 		Log(Error, "Failed to get channel group members \n%v \n%v", res, err)
// 		return res, err
// 	}

// 	var successful int = 0
// 	var attempted int = 0

// 	for _, user := range body {
// 		for i := 0; i < len(user.ActiveSessionIds); i++ {
// 			res, err := TSClient.UserPoke(user.ActiveSessionIds[i], msg)
// 			if err != nil {
// 				Log(Error, "Failed to poke %v \n%v \n%v", user.Nickname, res, err)
// 			}

// 			// Increase counters
// 			attempted++
// 			if res.IsSuccess {
// 				successful++
// 			}
// 		}
// 	}

// 	return &QueryResponse{
// 		Id:        -1,
// 		Msg:       fmt.Sprintf("%v out of %v clients sucesfully poked", successful, attempted),
// 		IsSuccess: true,
// 	}, nil
// }
