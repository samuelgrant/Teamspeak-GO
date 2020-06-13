package main

import (
	"fmt"
	"strconv"
	"strings"
)

type PrivilegeKey struct {
	ChannelId    int64
	Description  string
	GroupId      int64
	Token        string
	Type         string
	CustomFields map[string]string
}

// Create a privilege key. The groupId is a server group id.
// CustomFields can be used to create unique IDs for a user. You can search for users by these IDs later
func (this *Conn) TokensAdd(groupId int, description string, customFields map[string]string) (QueryResponse, PrivilegeKey, error) {
	s := fmt.Sprintf("tokenadd tokentype=0 tokenid1=%v tokenid2=0 tokendescription=%v", groupId, Encode(description))

	if len(customFields) > 0 {
		s = fmt.Sprintf("%v tokencustomset=", s)

		str := ""

		for k, v := range customFields {
			kv := fmt.Sprintf("ident=%v\\svalue=%v",
				strings.ReplaceAll(k, " ", "_"),
				strings.ReplaceAll(v, " ", "_"),
			)

			str = fmt.Sprintf("%v\\p%v", str, kv)
		}

		s = fmt.Sprintf("%v%v", s, strings.TrimLeft(str, "\\p"))

	}

	res, err := this.Exec(s)
	if err != nil {
		return ParseQueryResponse(res), PrivilegeKey{}, nil
	}

	parts := strings.Split(res, "\n")
	token, err := ParsePrivilegeKey(parts[0])
	if err != nil {
		return QueryResponse{}, PrivilegeKey{}, nil
	}
	return ParseQueryResponse(parts[1]), token, nil
}

// Delete a privilege key from the server
func (this *Conn) TokensDelete(token string) (QueryResponse, error) {

	res, err := this.Exec("privilegekeydelete token=%v", token)
	if err != nil {
		return ParseQueryResponse(res), err
	}

	return ParseQueryResponse(res), nil
}

// https://lethallhost.com.br/clientes/TeamSpeak%203%20Server%20Query%20Manual.pdf#page=32&zoom=100,72,306
func (this *Conn) Tokenslist() (QueryResponse, []PrivilegeKey, error) { //[]tokens  as well
	var Tokens []PrivilegeKey

	res, err := this.Exec("privilegekeylist")
	if err != nil {
		return QueryResponse{}, Tokens, err
	}

	keys := strings.Split(strings.Split(res, "\n")[0], "|")
	for i := 0; i < len(keys); i++ {
		token, err := ParsePrivilegeKey(keys[i])
		if err != nil {
			return ParseQueryResponse(strings.Split(res, "\n")[1]), nil, err
		}

		Tokens = append(Tokens, token)
	}

	return ParseQueryResponse(strings.Split(res, "\n")[1]), Tokens, nil
}

func ParsePrivilegeKey(s string) (PrivilegeKey, error) {
	parts := strings.Split(s, " ")
	token := PrivilegeKey{}

	token.Token = GetVal(parts[0])

	if len(parts) > 1 {
		if GetVal(parts[1]) == "0" {
			token.Type = "server"

			groupId, err := strconv.ParseInt(GetVal(parts[2]), 10, 64)
			if err != nil {
				return PrivilegeKey{}, err
			}
			token.GroupId = groupId
			token.ChannelId = 0
		} else {
			token.Type = "channel"

			channelId, err := strconv.ParseInt(GetVal(parts[2]), 10, 64)
			if err != nil {
				return PrivilegeKey{}, nil
			}
			token.ChannelId = channelId

			groupId, err := strconv.ParseInt(GetVal(parts[3]), 10, 64)
			if err != nil {
				return PrivilegeKey{}, nil
			}
			token.GroupId = groupId
		}

		token.Description = Decode(GetVal(parts[5]))
		token.CustomFields = parseCustomSets(parts[6])
	}

	return token, nil
}

func parseCustomSets(s string) map[string]string {
	CustomSets := make(map[string]string)

	s = strings.ReplaceAll(s, "token_customset=", "")
	parts := strings.Split(s, "\\p")
	for i := 0; i < len(parts); i++ {
		p := strings.Split(parts[i], "\\s")

		CustomSets[strings.ReplaceAll(GetVal(p[0]), "_", " ")] = strings.ReplaceAll(GetVal(p[1]), "_", " ")
	}

	return CustomSets
}
