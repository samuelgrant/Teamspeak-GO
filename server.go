package ts3

import (
	"encoding/json"
)

type VirtualServer struct {
	Id            int64  `json:"virtualserver_id,string"`
	Port          int64  `json:"virtualserver_port,string"`
	Status        string `json:"virtualserver_status"`
	ClientsOnline int64  `json:"virtualserver_clientsonline,string"`
	MaxClients    int64  `json:"virtualserver_maxclients,string"`
	Uptime        int64  `json:"virtualserver_uptime,string"`
	Name          string `json:"virtualserver_name"`
	Autostart     bool   `json:"virtualserver_autostart"`
}

// Send a global message to the current server
func ServerGlobalMessage(msg string) (*status, error) {
	queries := []KeyValue{
		{key: "msg", value: msg},
	}

	qres, _, err := get("gm", false, queries)
	if err != nil || !qres.IsSuccess() {
		Log(Error, "Failed to send global message \n%v\n%v", qres, err)
	}

	return qres, err
}

// Start a virtual server
func ServerStart(sid int64) (*status, error) {
	queries := []KeyValue{
		{key: "sid", value: i64tostr(sid)},
	}

	qres, _, err := get("serverstart", true, queries)
	if err != nil || !qres.IsSuccess() {
		Log(Error, "Failed to start server %v \n%v\n%v", sid, qres, err)
	}

	return qres, err
}

// Stop a virtual server
func ServerStop(sid int64) (*status, error) {
	queries := []KeyValue{
		{key: "sid", value: i64tostr(sid)},
	}

	qres, _, err := get("serverstop", true, queries)
	if err != nil || !qres.IsSuccess() {
		Log(Error, "Failed to stap server %v \n%v\n%v", sid, qres, err)
	}

	return qres, err
}

// List all virtual servers
func ServersList() (*status, []VirtualServer, error) {
	qres, body, err := get("serverlist", true)
	if err != nil || !qres.IsSuccess() {
		Log(Error, "Failed to get a list of virtual servers \n%v\n%v", qres, err)
		return qres, nil, err
	}

	var servers []VirtualServer
	json.Unmarshal([]byte(body), &servers)

	return qres, servers, err
}
