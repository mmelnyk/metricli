package metricer

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"

	"golang.org/x/net/http2"
)

type client struct {
	host      string
	transport *http.Transport
	timeout   time.Duration
}

// NewClient return client metrigo client
func NewClient(host string) *client {
	transport := &http.Transport{
		Dial: (&net.Dialer{
			Timeout:   100 * time.Second,
			KeepAlive: 30 * time.Second,
		}).Dial,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 0,
	}

	if err := http2.ConfigureTransport(transport); err != nil {
		fmt.Println("HTTP/2 is disabled: ", err)
	}

	return &client{host: "http://" + host, transport: transport, timeout: 100 * time.Second}
}

func (c *client) request(method, path string, data interface{}) (res json.RawMessage, latency time.Duration, err error) {
	httpclient := &http.Client{
		Transport: c.transport,
		Timeout:   c.timeout,
	}

	// Request body
	var reqbody io.Reader

	// Pass some data to the server?
	if data != nil {
		var reqdata []byte
		reqdata, err = json.Marshal(data)
		if err != nil {
			return
		}
		reqbody = bytes.NewBuffer(reqdata)
	}

	request, err := http.NewRequest(method, c.host+path, reqbody)
	if err != nil {
		return
	}

	request.Header.Set("Accept", "application/json")

	if data != nil {
		request.Header.Set("Content-Type", "application/json; charset=utf-8")
	}

	starttime := time.Now()
	response, err := httpclient.Do(request)
	if err != nil {
		return
	}
	defer response.Body.Close()

	// Read content of response body first
	content, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return
	}
	latency = time.Since(starttime)

	// Validation process
	contenttype := response.Header.Get("Content-Type")
	if i := strings.IndexRune(contenttype, ';'); i != -1 {
		contenttype = contenttype[0:i]
	}
	if contenttype != "application/json" {
		err = errors.New("Unexpected content-type value")
		return
	}

	var responseObject json.RawMessage

	if err = json.Unmarshal(content, &responseObject); err != nil {
		return
	}

	var apierror ApiError
	if err = json.Unmarshal(responseObject, &apierror); err != nil {
		return
	}

	if apierror.Error.Code != 0 {
		err = errors.New(apierror.Error.Message)
		return
	}

	// Check status code - should be OK(200) for normal response
	if response.StatusCode != http.StatusOK {
		err = errors.New(response.Status)
		return
	}

	res = responseObject

	return
}

func (c *client) requestBlob(method, path string) (res []byte, latency time.Duration, err error) {
	httpclient := &http.Client{
		Transport: c.transport,
		Timeout:   c.timeout,
	}

	request, err := http.NewRequest(method, c.host+path, nil)
	if err != nil {
		return
	}

	starttime := time.Now()
	response, err := httpclient.Do(request)
	if err != nil {
		return
	}
	// Read content of response body first
	content, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return
	}
	latency = time.Since(starttime)

	response.Body.Close()

	// Check status code - should be OK(200) for normal response
	if response.StatusCode != http.StatusOK {
		err = errors.New(response.Status)
		return
	}

	res = content

	return
}

func (c *client) HealthCheck() (*Health, time.Duration, error) {
	health := &Health{Status: "failed"}
	res, latency, err := c.request("GET", "/health/check", nil)
	if err != nil {
		return health, latency, err
	}

	if err = json.Unmarshal(res, health); err != nil {
		return health, latency, err
	}

	return health, latency, nil
}

func (c *client) GetMetrics() (RawMetrics, time.Duration, error) {
	var metrics RawMetrics
	res, latency, err := c.request("GET", "/metrics/values", nil)
	if err != nil {
		return metrics, latency, err
	}

	decoder := json.NewDecoder(bytes.NewReader(res))
	decoder.UseNumber()

	if err = decoder.Decode(&metrics); err != nil {
		return metrics, latency, err
	}

	return metrics, latency, nil
}

func (c *client) GetLogLevels() (LogLevels, time.Duration, error) {
	var levels LogLevels
	res, latency, err := c.request("GET", "/debug/logger/levels", nil)
	if err != nil {
		return levels, latency, err
	}

	decoder := json.NewDecoder(bytes.NewReader(res))
	if err = decoder.Decode(&levels); err != nil {
		return levels, latency, err
	}

	return levels, latency, nil
}

/*
func (c *client) SetLogLevels(logger, level string) (time.Duration, error) {
	path := fmt.Sprintf("/debug/logger/level/%s/%s", logger, level)
	_, latency, err := c.request("PUT", path, nil)
	return latency, err
}
*/

func (c *client) GetTrace() ([]byte, time.Duration, error) {
	res, latency, err := c.requestBlob("GET", "/debug/pprof/trace?seconds=5")
	return res, latency, err
}

func (c *client) GetProfile() ([]byte, time.Duration, error) {
	res, latency, err := c.requestBlob("GET", "/debug/pprof/profile")
	return res, latency, err
}

func (c *client) GetJsonMetrics() ([]byte, time.Duration, error) {
	res, latency, err := c.request("GET", "/metrics/values", nil)
	return res, latency, err
}

func (c *client) GetTextMetrics() ([]byte, time.Duration, error) {
	res, latency, err := c.requestBlob("GET", "/metrics/values")
	return res, latency, err
}

func (c *client) GetBlock() ([]byte, time.Duration, error) {
	res, latency, err := c.requestBlob("GET", "/debug/pprof/block")
	return res, latency, err
}

func (c *client) GetGoroutine() ([]byte, time.Duration, error) {
	res, latency, err := c.requestBlob("GET", "/debug/pprof/goroutine")
	return res, latency, err
}

func (c *client) GetHeap() ([]byte, time.Duration, error) {
	res, latency, err := c.requestBlob("GET", "/debug/pprof/heap")
	return res, latency, err
}

func (c *client) GetMutex() ([]byte, time.Duration, error) {
	res, latency, err := c.requestBlob("GET", "/debug/pprof/mutex")
	return res, latency, err
}

func (c *client) GetThreadCreate() ([]byte, time.Duration, error) {
	res, latency, err := c.requestBlob("GET", "/debug/pprof/threadcreate")
	return res, latency, err
}
