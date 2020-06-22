# A simple TS3 Server Query API in GO
Written for the Eve Online alliance Boom & because I felt like learning GO...

This library is in early development and does not provide full coverage for all Team Speak Query Commands. The full query command reference can be [found here](./docs/TeamSpeak%203%20Server%20Query%20Manual.pdf). If you want to execute a command that has not been implemented in the library, you can use the Get function (see below). Documentation for the implemented commands can be found in the [wiki](https://github.com/samuelgrant/Teamspeak-GO/wiki).

## Issues and Feature requests
While this library has been developed for a specific group, feel free to use, fork and open pull requests or open issues to report bugs or request features.

## Configuration
Unlike the old server query which uses Telnet, this library uses the new WebQuery over HTTP/HTTPS. You will need a Team Speak server running version **3.12.0** or later. Start the server with the command `./ts3server query_protocols=raw,ssh,http,https`.

<details>
    <summary>Docker Configuration</summary>

```yml
version: '3'
services:
    ts3:
        image: teamspeak
        command: ['ts3server', 'query_protocols=raw,ssh,http,https']
        ports:
            - 9987:9987/udp
            - 10011:10011
            - 30033:30033
            - 10080:10080
        environment:
            TS3SERVER_LICENSE: accept
```

</details>

### Creating an API Key
The WebQuery requires API calls to be sent with an authentication header `x-api-key: {api token}`. While API tokens must have one of three scopes (outlined below), this library requires the **manage** scope to function correctly. New API keys have a lifetime of 14 days, but you can override this using the `lifetime=` parameter, using `lifetime=0` will stop the API key from expiring.

**Scopes:**
* manage
* read, or,
* write

**Creating an API Token using Telnet Server Query**

```
Telnet <server address> <port> //localhost 10011
login <query username> <query password> // Use the query account generated when the server ran for the first time
apikeyadd scope=manage lifetime=0
```
**Example Output**
```
apikey=<api token> id=4 sid=0 cldbid=1 scope=manage time_left=unlimited created_at=1592602125 expires_at=1592602125
error id=0 msg=ok
```

## Simple Usage
The example below assumes your code is in the main function of main.go
```golang
  // These variables should come from environment variables and not be in your source code
  host, apikey := "localhost:10080": "<api key>"

  // Logging is disabled by default
  ts3.LoggingEnabled(true)

  // This needs to be done before you can make API calls
  ts3.ConfigureHttp(apiKey, host)

  // The default virtual server is 1
  // Override that at any time using the function below
  ts3.SelectVirtualServer(1)

  // We can discard all results of a function
  ts3.ServerGlobalMessage("Global text message sent to the server")
  // or we can grab the results of a function and handle errors/results
  qres, servers, err := ts3.ServersList()
  if err != nil || !qres.IsSuccess() {
    // handle error
  }

```
For a full list of supported command and a description on response types please view the wiki.

### Old Example (Old) - Prior to v1.0.0-alpha

<details>
	<summary>Applies to releases before v1.0.0-alpha</summary>

The example below assumes your code is in the main function of `main.go`
```golang
package main

import (
	"log"

	ts3 "github.com/samuelgrant/Teamspeak-GO"
)

func main() {
  // These variables should come from environment variables and not be in your source code
  address, port, username, password := "localhost": "10011", "<query username>", "<query user password>"

  // Library logging is disabled by default, you can change this setting at any time using the function below
  ts3.LoggingEnabled(true)

  // Connect the library to your Team Speak server and get a connection object
  client, err := ts3.Connect(address + ":" + port)
  if err != nil {
    // Handle errors here.
    // If an error has occurred here you cannot go any further until you fix the problem
  }

  // Login to the server
  err = client.Login(username, password)
  if err != nil {
    // Handle errors here.
    // If you got this far you connected to the server, but you were not able to login. Check you have the correct login credentials for a server query account
  }

  // Select a virtual server - Every command except for Start, Stop, ServerList and Use require that you have selected a virtual server
  // You can change the selected virtual server at any time using the command below
  qres, err := client.Use(sid)//Virtual Server Id
  if err != nil || !qres.IsSuccess {
    // Handle errors here.
    // Most commands return a queryResponse struct. If an err is not returned but the command fails the query response (qres) will have a message explaining why your command was rejected
  }

  // Example of a library command
  // The Team Speak server query requires that parameters with spaces, tabs and new lines be escaped.
  // Library functions will handle this for you. If you are creating your own command you can do that using the `Escape()` function
  res, err = client.GlobalMessage("Message string")

  // Using Exec to fire a custom command
  // The exec function will always return a second parameter of type string.
  // Depending on the response from Team Speak this could be an empty string or data requested from the server.
  // You can discard this value by using an underscore `res, _, err :=`
  res, msg, err := client.Exec("clientmove clid=%v cid=%v", 7, 13)
}
```
</details>