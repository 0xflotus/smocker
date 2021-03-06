package types

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	log "github.com/sirupsen/logrus"
)

type History []*Entry

type Sessions []*Session

type Session struct {
	ID      string    `json:"id"`
	Name    string    `json:"name"`
	Date    time.Time `json:"date"`
	History History   `json:"history"`
	Mocks   Mocks     `json:"mocks"`
}

type Entry struct {
	MockID   string   `json:"mock_id,omitempty"`
	Request  Request  `json:"request"`
	Response Response `json:"response"`
}

type Request struct {
	Path        string      `json:"path"`
	Method      string      `json:"method"`
	Body        interface{} `json:"body,omitempty" yaml:"body,omitempty"`
	BodyString  string      `json:"-" yaml:"-"`
	QueryParams url.Values  `json:"query_params,omitempty" yaml:"query_params,omitempty"`
	Headers     http.Header `json:"headers,omitempty" yaml:"headers,omitempty"`
	Date        time.Time   `json:"date" yaml:"date"`
}

type Response struct {
	Status  int         `json:"status"`
	Body    interface{} `json:"body,omitempty" yaml:"body,omitempty"`
	Headers http.Header `json:"headers,omitempty" yaml:"headers,omitempty"`
	Date    time.Time   `json:"date" yaml:"date"`
}

func HTTPRequestToRequest(req *http.Request) Request {
	bodyBytes := []byte{}
	if req.Body != nil {
		var err error
		bodyBytes, err = ioutil.ReadAll(req.Body)
		if err != nil {
			log.WithError(err).Error("Failed to read request body")
		}
	}
	req.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
	var body interface{}
	var tmp map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &tmp); err != nil {
		body = string(bodyBytes)
	} else {
		body = tmp
	}

	headers := http.Header{}
	for key, values := range req.Header {
		headers[key] = make([]string, 0, len(values))
		for _, value := range values {
			headers.Add(key, value)
		}
	}
	headers.Add("Host", req.Host)

	return Request{
		Path:        req.URL.Path,
		Method:      req.Method,
		Body:        body,
		BodyString:  string(bodyBytes),
		QueryParams: req.URL.Query(),
		Headers:     headers,
		Date:        time.Now(),
	}
}
