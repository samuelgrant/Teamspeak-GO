package ts3

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
)

type TokenType int

const (
	Server  = 0
	Channel = 1
)

type PrivilegeKey struct {
	ChannelId    int64  // `json:"virtualserver_id,string"`
	Description  string `json:"token_description,string"`
	GroupId      int64
	Token        string    `json:"token"`
	Type         TokenType `json:"token_type,string"`
	CustomFields map[string]string
}

// Create a privilege key. The groupId is a server group id.
// CustomFields can be used to add information to a DbUser such as an ID from an external authentication provider
// Users can be searched for using custom fields
func TokensAdd(sgid int64, description string, customFields map[string]string) (*status, *PrivilegeKey, error) {
	// Build custom fields
	str := ""
	for k, v := range customFields {
		kv := fmt.Sprintf("ident=%v value=%v",
			strings.ReplaceAll(k, " ", "_"),
			strings.ReplaceAll(v, " ", "_"),
		)

		str = fmt.Sprintf("%v%v|", str, kv)
	}

	queries := []KeyValue{
		{key: "tokentype", value: "0"},
		{key: "tokenid1", value: i64tostr(sgid)},
		{key: "tokenid2", value: "0"},
		{key: "tokendescription", value: description},
		{key: "tokencustomset", value: strings.TrimRight(str, "|")},
	}

	qres, body, err := get("tokenadd", false, queries)
	if err != nil || !qres.IsSuccess() {
		Log(Error, "Failed to create a new privilege token \n%v\n%v", qres, err)
		return qres, nil, err
	}

	token := []PrivilegeKey{}
	json.Unmarshal([]byte(body), &token)

	// Setup other fields
	token[0].Description = description
	token[0].GroupId = sgid
	token[0].Type = Server
	token[0].CustomFields = customFields

	return qres, &token[0], err
}

// Delete a privilege key from the server
func TokensDelete(token string) (*status, error) {
	queries := []KeyValue{
		{key: "token", value: token},
	}

	qres, _, err := get("privilegekeydelete", false, queries)
	if err != nil || !qres.IsSuccess() {
		Log(Error, "Failed to delete privilegekey %v \n%v\n%v", token, qres, err)
	}

	return qres, err
}

// List active privilege keys, include their custom field sets
func TokensList() (*status, []PrivilegeKey, error) {
	var PrivilegeKeys []PrivilegeKey

	qres, body, err := get("privilegekeylist", false)
	if err != nil || !qres.IsSuccess() {
		Log(Error, "Failed to get a list of privilege keys (tokens) \n%v\n%v", qres, err)
		return qres, nil, err
	}

	// get each token as an item
	var lines []json.RawMessage
	json.Unmarshal([]byte(body), &lines)

	for _, line := range lines {
		var token PrivilegeKey
		json.Unmarshal(line, &token)
		PrivilegeKeys = append(PrivilegeKeys, token)
	}

	return qres, PrivilegeKeys, err
}

// Build a privilege key struct from a string
func (p *PrivilegeKey) UnmarshalJSON(data []byte) error {
	// We need to Unmarshal the JSON string into a map
	// we will hold the map data in 'v' so we can access it
	var v map[string]string
	if err := json.Unmarshal(data, &v); err != nil {
		log.Fatal(err)
		return err
	}

	p.Token = v["token"]

	// Escape this function if we only need to grab the token
	// this occurs on the AddTokens function
	if len(v) <= 1 {
		return nil
	}

	// Get the tokens as int64s
	tokenId1, err := strconv.ParseInt(v["token_id1"], 10, 64)
	tokenId2, err := strconv.ParseInt(v["token_id2"], 10, 64)
	if err != nil {
		Log(Error, "Failed to parse privilege key %v \n%v", v["token"], err)
		return err
	}

	// Parse custom fields and build up the map
	customFields := make(map[string]string)
	customFieldSets := strings.Split(v["token_customset"], " ")
	for _, set := range customFieldSets {
		kvp := strings.Split(set, "=")

		customFields[kvp[0]] = kvp[1]
	}

	// Build up the token
	p.CustomFields = customFields
	p.Description = v["token_description"]
	if v["token_type"] == "0" {
		p.Type = Server
		p.GroupId = tokenId1
		p.ChannelId = -1
	} else {
		p.Type = Channel
		p.ChannelId = tokenId1
		p.GroupId = tokenId2
	}

	return nil
}
