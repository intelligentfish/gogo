package restClient

import (
	"fmt"
	"net/http"
	"time"

	"github.com/valyala/fasthttp"
)

const (
	DefaultHost = "localhost"
	DefaultPort = 9200
)

// RestClient rest client
type RestClient struct {
	host string // es host
	port int    // es port
}

// Use use options
func (object *RestClient) Use(options ...Option) *RestClient {
	for _, option := range options {
		option(object)
	}
	if 0 >= len(object.host) {
		object.host = DefaultHost
	}
	if 0 >= object.port {
		object.port = DefaultPort
	}
	return object
}

// Do do command in timeout
func (object *RestClient) Do(cmd CMD, timeout time.Duration) (err error) {
	req := fasthttp.AcquireRequest()
	res := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(res)
	req.SetRequestURI(fmt.Sprintf("http://%s:%d%s",
		object.host,
		object.port,
		cmd.Uri()))
	req.Header.SetMethod(cmd.HTTPMethod())
	req.Header.SetContentType("application/json; charset=utf-8")
	if err = cmd.SetRequestBody(req.BodyWriter()); nil != err {
		return
	}
	if err = fasthttp.DoTimeout(req, res, timeout); nil != err {
		return
	}
	switch res.StatusCode() {
	case http.StatusOK, http.StatusCreated:
		err = cmd.ProcessResponseBody(res.Body())
	default:
		err = fmt.Errorf("(%d,%s)", res.StatusCode(), string(res.Body()))
	}
	return
}

// NewRestClient factory method
func NewRestClient(options ...Option) *RestClient {
	object := &RestClient{}
	return object.Use(options...)
}
