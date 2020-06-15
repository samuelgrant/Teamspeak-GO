package main

import (
	"strconv"
	"strings"
)

type VirtualServer struct {
	Id                 int64
	Port               int64
	Status             string
	ClientsOnline      int64
	QueryClientsOnline int64
	MaxClients         int64
	Uptime             int64
	Name               string
	Autostart          bool
}

// Select a virtual server to manage
func (this *Conn) Use(sid int) (*QueryResponse, error) {
	res, _, err := this.Exec("use %v", sid)
	if err != nil || !res.IsSuccess {
		Log(Error, "Failed to select virtual server %v.\n%v\n%v", sid, res, err)
		return res, err
	}

	return res, nil
}

// Send a global message to the current server
func (this *Conn) GlobalMessage(msg string) (*QueryResponse, error) {
	res, _, err := this.Exec("gm msg=%v", Encode(msg))
	if err != nil || !res.IsSuccess {
		Log(Error, "Failed to send a global message %v.\n%v\n%v", msg, res, err)
		return res, err
	}

	return res, nil
}

// Start a virtual server
func (this *Conn) Start(sid int) (*QueryResponse, error) {
	res, _, err := this.Exec("serverstart sid=%v", sid)
	if err != nil || !res.IsSuccess {
		Log(Error, "Failed to start virtual server %v.\n%v\n%v", sid, res, err)
		return res, err
	}

	return res, nil
}

// Stop a virtual server
func (this *Conn) Stop(sid int) (*QueryResponse, error) {
	res, _, err := this.Exec("serverstop sid=%v", sid)
	if err != nil || !res.IsSuccess {
		Log(Error, "Failed to stop virtual server %v.\n%v\n%v", sid, res, err)
		return res, err
	}

	return res, nil
}

// List all virtual servers
func (this *Conn) List() (*QueryResponse, []VirtualServer, error) {
	var ServerList []VirtualServer

	res, body, err := this.Exec("serverlist")
	if err != nil || !res.IsSuccess {
		Log(Error, "Failed to get virtual servers from Team Speak.\n%v\n%v", res, err)
		return res, ServerList, err
	}

	servers := strings.Split(body, "|")
	for i := 0; i < len(servers); i++ {
		server, err := ParseVirtualServer(servers[i])
		if err != nil {
			Log(Error, "Failed to parse server information.\n%v\n%v", res, err)
			return res, nil, err
		}

		ServerList = append(ServerList, server)
	}

	return res, ServerList, err
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
