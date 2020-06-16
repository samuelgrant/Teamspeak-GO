package ts3

import (
	"fmt"
	"io"
	"net"
	"strings"
	"time"
)

const (
	DefaultPort = 10011
	Bytes       = 64 * 1024
)

var (
	DialTimeout = 20 * time.Second
)

type Conn struct {
	conn net.Conn
}

// Attempt to authenticate with the server
func (this *Conn) Login(user string, passwd string) error {
	_, _, err := this.Exec("login %v %v", Encode(user), Encode(passwd))
	return err
}

// Open a connection to the server
func Connect(address string) (*Conn, error) {
	var (
		err error
	)

	// Ensure we have a query port. If no port is provided
	// then we shall use the default query port
	strings.TrimSpace(address)
	if !strings.Contains(address, ":") {
		address = fmt.Sprintf("%v:%v", address, DefaultPort)
	}

	// Establish a connection
	conn, err := net.DialTimeout("tcp", address, DialTimeout)
	if err != nil {
		Log(Error, "Failed to establish a TCP connection to the server")
		return nil, err
	}

	// Throw away this response
	conn.Read(make([]byte, Bytes))

	Log(Notice, "Successfully established a TCP connection to %v", address)

	return &Conn{
		conn: conn,
	}, nil
}

// Close the TCP connection
func (TSClient *Conn) Disconnect() error {
	_, _, err := TSClient.Exec("quit")
	if err != nil {
		Log(Error, "Failed to disconnect from the Team Speak server @ %v", TSClient.conn.RemoteAddr().String)
	}

	Log(Notice, "Closing TCP conncetion")

	return TSClient.conn.Close()
}

// Checks if the TCP client is connected to the server
func (this *Conn) IsConnected() bool {
	one := make([]byte, 1)

	res, err := this.conn.Read(one)
	if res == 0 || err == io.EOF {
		this.conn.Close()
		return false
	}

	return true
}

// Send a command to the server
func (this *Conn) Exec(format string, a ...interface{}) (*QueryResponse, string, error) {
	// Response Object
	res := make([]byte, Bytes)

	// Generate the string
	cmd := fmt.Sprintf(format+"\n", a...)

	// Send the request
	_, err := this.conn.Write([]byte(cmd))
	if err != nil {
		Log(Error, "Failed to send the command to the server\n  Command: %v\n  Error:%v", cmd, err)
		return nil, "", err
	}

	// Get the response from the server
	_, err = this.conn.Read(res)
	if err != nil {
		Log(Error, "Failed to get a response from the server\n  Res: %v\n  Error:", string(res), err)
		return nil, "", err
	}

	// Return result
	Log(CmdExc, cmd)

	return ParseQueryResponse(string(res))
}
