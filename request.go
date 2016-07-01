package restclient

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
)

type RequestBuilder interface {
	NewRequest(*Config) error
	Process() error
}

//The REST API request
type Request struct {
	Config       *Config
	Operation    *Operation
	HTTPRequest  *http.Request
	HTTPResponse *http.Response
	StatusCode   int
}

func BuildRequest(c *Config, o *Operation) (r *Request, err error) {
	// Set path to root if empty and add root slash to path is missing from the start
	p := o.httpPath
	if p == "" {
		p = "/"
	} else if p[0:1] != "/" {
		p = "/" + p
	}
	// Remove trailing slash from endpoint URL
	e := *c.EndPoint
	if e[len(e)-1:] == "/" {
		e = e[0 : len(e)-1]
	}
	// Only accept certain methods
	var method string
	switch o.httpMethod {
	case "GET":
		method = "GET"
	case "POST":
		method = "POST"
	case "PUT":
		method = "PUT"
	case "PATCH":
		method = "PATCH"
	default:
		method = "GET"
	}
	HTTPReq, err := http.NewRequest(method, "", bytes.NewBuffer(o.sendData))
	if err != nil {
		return
	}
	HTTPReq.URL, err = url.Parse(e + p)
	if err != nil {
		return
	}
	HTTPReq.Close = true
	if c.UserId != nil {
		if c.Password == nil {
			var password string
			c.Password = &password
		}
		HTTPReq.SetBasicAuth(*c.UserId, *c.Password)
	}

	r = &Request{
		Config:      c,
		Operation:   o,
		HTTPRequest: HTTPReq,
	}
	return
}

func Send(r *Request) (httpCode *int, err error) {
	r.HTTPResponse, err = r.Config.HTTPClient.Do(r.HTTPRequest)
	if err != nil {
		code := http.StatusServiceUnavailable
		httpCode = &code
		return
	}
	r.StatusCode = r.HTTPResponse.StatusCode
	httpCode = &r.StatusCode

	defer r.HTTPResponse.Body.Close()
	var dec *json.Decoder
	if r.HTTPResponse.ContentLength > 0 {
		dec = json.NewDecoder(io.LimitReader(r.HTTPResponse.Body, r.HTTPResponse.ContentLength))
	} else {
		dec = json.NewDecoder(r.HTTPResponse.Body)
	}
	dec.Decode(r.Operation.responsePtr)
	return
}
