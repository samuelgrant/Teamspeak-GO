package ts3

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

var (
	api_           string
	baseUrl_       string
	virtualServer_ int    = 1
	scheme_        string = "http"
)

type Response struct {
	Status status          `json:"status"`
	Body   json.RawMessage `json:"body"`
}

type status struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (s status) IsSuccess() bool {
	return s.Message == "ok" || s.Code == -1
}

type KeyValue struct {
	key   string
	value string
}

// Adjust the HTTP Settings
func ConfigureHttp(apiKey, baseUrl string, useHttps bool) {
	api_ = apiKey
	baseUrl_ = baseUrl
	if useHttps {
		scheme_ = "https"
	} else {
		scheme_ = "http"
	}
	Log(Notice, "HTTP Config set")
}

// Select a virtual server (defaults to 1)
func SelectVirtualServer(sid int) {
	virtualServer_ = sid
}

// HTTP Get request, taxes in optional []KeyValues which will be built as URL queries
func get(path string, globalCmd bool, queries ...[]KeyValue) (qres *status, body string, err error) {
	sid := ""
	if !globalCmd {
		sid = fmt.Sprintf("%v/", virtualServer_)
	}

	baseUrl, err := url.Parse(fmt.Sprintf("%v://%v/%v%v", scheme_, baseUrl_, sid, path))
	if err != nil {
		Log(Error, "Failed to build request URL \n%v", err)
		return nil, "", err
	}

	// Build the query parts
	params := url.Values{}
	if len(queries) > 0 {
		for _, kvp := range queries[0] {
			params.Add(kvp.key, kvp.value)
		}
	}
	baseUrl.RawQuery = params.Encode()

	req, err := http.NewRequest("GET", baseUrl.String(), nil)
	if err != nil {
		Log(Error, "Error building an HTTP request \n%v", err)
		return nil, "", err
	}

	// Exectue the request
	res, err := doRequest(req, err, &http.Client{})
	if err != nil {
		Log(Error, "Error executing HTTP request \n%v", err)
		return nil, "", err
	}

	var r Response
	json.Unmarshal(res, &r)
	return &r.Status, string(r.Body), err
}

// Exectue HTTP Request
func doRequest(req *http.Request, err error, client *http.Client) ([]byte, error) {
	req.Header.Set("x-api-key", api_)

	Log(CmdExc, "%v", req.URL)
	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}
