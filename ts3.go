package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"time"
)

const (
	DefaultPort = 10011
	Bytes       = 64 * 1024
)

var (
	DialTimeout = 1 * time.Second
)

type Conn struct {
	conn net.Conn
}

// Attempt to authenticate with the server
func (this *Conn) Login(user string, passwd string) error {
	_, err := this.Exec("login %v %v", Encode(user), Encode(passwd))
	return err
}

// Open a connection to the server
func Connect(address string) (*Conn, error) {
	var (
		err error
		// line string
	)

	/**
	* Ensure we have a query port.
	* IF no port is provided use the default query port
	 */
	strings.TrimSpace(address)
	if !strings.Contains(address, ":") {
		address = fmt.Sprintf("%v:%v", address, DefaultPort)
	}

	// Establish a connection
	conn, err := net.DialTimeout("tcp", address, DialTimeout)
	if err != nil {
		log.Printf(ErrorColor, fmt.Sprintf("[error]: Failed to establish a TCP connection to the server"))
		return nil, err
	}

	ts3conn := &Conn{
		conn: conn,
	}

	res := make([]byte, Bytes)

	// Get response & check we are connected to TS3
	_, err = conn.Read(res)
	// content := string(res)
	// if !strings.Contains(content.ReadLine(), "TS3") {
	// 	log.Printf("Connection is not a TeamSpeak 3 server")
	// 	return nil, errors.New("Connection is not a TeamSpeak 3 server")
	// }

	log.Printf(NoticeColor, fmt.Sprintf("[info]: Connected to the TS3 server @ %v", conn.RemoteAddr()))

	return ts3conn, nil
}

// Close the TCP connection
func (this *Conn) Disconnect() error {
	_, err := this.Exec("quit")
	if err != nil {
		log.Printf(ErrorColor, "[error]: Failed to disconnect from the teamspeak")
	}

	log.Printf(NoticeColor, "Closing TCP Connection")
	return this.conn.Close()
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
func (this *Conn) Exec(format string, a ...interface{}) (string, error) {
	// Response Object
	reply := make([]byte, Bytes)

	// Generate the string
	s := fmt.Sprintf(format+"\n", a...)

	// Send the request
	_, err := this.conn.Write([]byte(s))
	if err != nil {
		log.Printf(ErrorColor, fmt.Sprintf("[error]: Failed to send command to the server\n%v\n%v", s, err))
		return "", nil
	}

	// Get the response from the server
	_, err = this.conn.Read(reply)
	if err != nil {
		log.Printf(ErrorColor, fmt.Sprintf("[error]: Failed to get server response\n%v\n%v", string(reply), err))
		return "", nil
	}

	log.Printf(NoticeColor, fmt.Sprintf("[info]: Executed Command: %v", s))
	return string(reply), nil
}
