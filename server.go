package main

import (
	"fmt"
	"log"
	"strings"
)

/**


TS3Client.Server.Disconnect()


TS3Client.Server.Info(sid) - Return { online int, started_at time, name string, version string, avaliable_slots, port int}


// TS3Client.Server.Use(sid)
// TS3Client.Server.GlobalMessage(message)
// TS3Client.Server.Start(sid)
// TS3Client.Server.Stop(sid)
// TS3Client.Server.List()
*/

// Select a virtual server to manage
func (this *Conn) Use(sid int) (QueryResponse, error) {
	qr, err := this.Exec("use %v", sid)
	if err != nil {
		return ParseQueryResponse(qr), err
	}

	log.Printf(InfoColor, fmt.Sprintf("[info]: Virtual server %v selected", sid))
	return ParseQueryResponse(qr), nil
}

func (this *Conn) GlobalMessage(msg string) (QueryResponse, error) {
	qr, err := this.Exec("gm msg=%v", Encode(msg))
	if err != nil {
		return ParseQueryResponse(qr), err
	}

	return ParseQueryResponse(qr), nil
}

func (this *Conn) Start(sid int) (QueryResponse, error) {
	qr, err := this.Exec("serverstart sid=%v", sid)
	if err != nil {
		return ParseQueryResponse(qr), err
	}

	return ParseQueryResponse(qr), nil
}

func (this *Conn) Stop(sid int) (QueryResponse, error) {
	qr, err := this.Exec("serverstop sid=%v", sid)
	if err != nil {
		return ParseQueryResponse(qr), nil
	}

	return ParseQueryResponse(qr), nil
}

func (this *Conn) List() ([]VirtualServer, error) {
	res, err := this.Exec("serverlist")

	var ServerList []VirtualServer

	if err != nil {
		return ServerList, err
	}

	servers := strings.Split(res, "|")

	for i := 0; i < len(servers); i++ {
		server, err := ParseVirtualServer(servers[i])
		if err != nil {
			return ServerList, err
		}

		ServerList = append(ServerList, server)
	}

	return ServerList, nil
}

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
