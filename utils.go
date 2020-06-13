package main

import (
	"strconv"
	"strings"
)

var (
	// encoder performs white space and special character encoding
	// as required by the ServerQuery protocol.
	encoder = strings.NewReplacer(
		`\`, `\\`,
		`/`, `\/`,
		` `, `\s`,
		`|`, `\p`,
		"\a", `\a`,
		"\b", `\b`,
		"\f", `\f`,
		"\n", `\n`,
		"\r", `\r`,
		"\t", `\t`,
		"\v", `\v`,
	)

	// decoder performs white space and special character decoding
	// as required by the ServerQuery protocol.
	decoder = strings.NewReplacer(
		`\\`, "\\",
		`\/`, "/",
		`\s`, " ",
		`\p`, "|",
		`\a`, "\a",
		`\b`, "\b",
		`\f`, "\f",
		`\n`, "\n",
		`\r`, "\r",
		`\t`, "\t",
		`\v`, "\v",
	)
)

// Decode and return the string
func Decode(s string) string {
	return decoder.Replace(s)
}

func Encode(s string) string {
	return encoder.Replace(s)
}

type QueryResponse struct {
	Id      int64
	Msg     string
	Success bool
}

// Parse a Teamspeak Server Query response
func ParseQueryResponse(s string) QueryResponse {

	parts := strings.Split(strings.TrimSpace(s), " ")

	Id, err := strconv.ParseInt(strings.Split(parts[1], "=")[1], 10, 64)
	if err != nil {
		Id = -1
	}

	message := strings.Split(parts[2], "=")[1]

	return QueryResponse{
		Id:      Id,
		Msg:     Decode(message),
		Success: message == "ok",
	}
}

// Parse a string into a Teamspeak Virtual Server object
func ParseVirtualServer(s string) (VirtualServer, error) {

	parts := strings.Split(strings.TrimSpace(s), " ")
	server := VirtualServer{}

	// Get the Virtual Server ID
	Id, err := strconv.ParseInt(strings.Split(parts[0], "=")[1], 10, 64)
	if err != nil {
		return VirtualServer{}, err
	}
	server.Id = Id

	// Get the virtual server portnumber
	Port, err := strconv.ParseInt(strings.Split(parts[1], "=")[1], 10, 64)
	if err != nil {
		return VirtualServer{}, err
	}
	server.Port = Port

	// Get the server status "online" || "offline"
	server.Status = strings.Split(parts[2], "=")[1]
	if server.Status == "offline" {
		server.Name = strings.Split(parts[3], "=")[1]
		server.Autostart = strings.Split(parts[4], "=")[1] == "1"

		return server, nil
	}

	// Get the number of clients online
	ClientsOnline, err := strconv.ParseInt(strings.Split(parts[3], "=")[1], 10, 64)
	if err != nil {
		return VirtualServer{}, err
	}
	server.ClientsOnline = ClientsOnline

	// Get the number of 'Query' clients online
	QueryClientsOnline, err := strconv.ParseInt(strings.Split(parts[4], "=")[1], 10, 64)
	if err != nil {
		return VirtualServer{}, err
	}
	server.QueryClientsOnline = QueryClientsOnline

	// Get the number of 'Client Slots' avaliable on the virtual server
	MaxClients, err := strconv.ParseInt(strings.Split(parts[5], "=")[1], 10, 64)
	if err != nil {
		return VirtualServer{}, err
	}
	server.MaxClients = MaxClients

	// Get the server uptime
	Uptime, err := strconv.ParseInt(strings.Split(parts[6], "=")[1], 10, 64)
	if err != nil {
		return VirtualServer{}, err
	}
	server.Uptime = Uptime

	server.Name = strings.Split(parts[7], "=")[1]
	server.Autostart = strings.Split(parts[4], "=")[1] == "8"

	return server, nil
}

func GetVal(s string) string {
	return strings.Split(s, "=")[1]
}
