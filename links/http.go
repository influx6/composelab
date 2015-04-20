package links

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"code.google.com/p/go-uuid/uuid"

	"github.com/influx6/composelab/arch"
)

//HTTPLink represents a standard servicelink using http underneath
//as its protocol transport
type HTTPLink struct {
	*arch.ServiceLink
	client *http.Client
}

//NewHTTPLink returns a new http service link
func NewHTTPLink(prefix string, addr string, port int) *HTTPLink {
	desc := arch.NewDescriptor("http", prefix, addr, port, "0", "http")
	return &HTTPLink{
		arch.NewServiceLink(desc),
		new(http.Client),
	}
}

//NewSecureHTTPLink returns a new http service link
func NewSecureHTTPLink(prefix string, addr string, port int, trans *http.Transport) *HTTPLink {
	cl := &http.Client{Transport: trans}
	desc := arch.NewDescriptor("http", prefix, addr, port, "0", "https")
	return &HTTPLink{
		arch.NewServiceLink(desc),
		cl,
	}
}

//NewHTTPWrap wraps a HTTPLink as a arch.Linkage
func NewHTTPWrap(h *HTTPLink) arch.Linkage {
	return arch.Linkage(h)
}

//Discover sends a request to the set server links if a service exists
func (hl *HTTPLink) Discover(target string, callback func(string, interface{}, interface{})) error {
	url := fmt.Sprintf("%s/%s", "discover", target)

	return hl.Request(url, target, nil, func(sets ...interface{}) {
		rq := sets[0]
		req, ok := rq.(*http.Request)

		// cb := sets[1]

		if !ok {
			return
		}

		req.Header.Set("X-Request-UUID", uuid.New())
		req.Header.Set("Content-Type", "application/json")

	}, func(rsd ...interface{}) {

		body, ok := rsd[0].([]byte)

		if !ok {
			log.Fatal("body not available", body)
			return
		}

		res, ok := rsd[1].(*http.Response)

		if !ok {
			log.Fatal("response object not available", res, ok)
			return
		}

		req, ok := rsd[2].(*http.Request)

		if !ok {
			log.Fatal("request object not available", req)
			return
		}

		status := res.StatusCode

		if status == 200 || status == 201 || status == 304 {

			jsn := new(arch.LinkDescriptor)
			err := json.Unmarshal(body, jsn)

			if err != nil {
				log.Fatal("json umarshalling error with /discover", err, req, res)
				return
			}

			callback(target, jsn, res)
		}
	})
}

//Register  registers a service to the specific server with the meta details as json
func (hl *HTTPLink) Register(target string, meta *arch.LinkDescriptor, cb func(d ...interface{})) error {
	jsn, err := json.Marshal(meta)
	// url := fmt.Sprintf("%s/%s", "register", target)

	if err != nil {
		return err
	}

	return hl.Request("register", target, bytes.NewReader(jsn), func(sets ...interface{}) {
		rq := sets[0]
		req, ok := rq.(*http.Request)

		// cb := sets[1]

		if !ok {
			return
		}

		req.Header.Set("X-Service-UUID", meta.UUID)
		req.Header.Set("X-Request-UUID", uuid.New())
		req.Header.Set("Content-Type", "application/json")

	}, func(resd ...interface{}) {
		cb(resd...)
	})

}

//Unregister  registers a service to the specific server with the meta details as json
func (hl *HTTPLink) Unregister(target string, meta *arch.LinkDescriptor, cb func(d ...interface{})) error {
	jsn, err := json.Marshal(meta)
	// url := fmt.Sprintf("%s/%s", "unregister", target)
	if err != nil {
		return err
	}

	return hl.Request("unregister", target, bytes.NewReader(jsn), func(sets ...interface{}) {
		rq := sets[0]
		req, ok := rq.(*http.Request)

		// cb := sets[1]

		if !ok {
			return
		}

		// req.Header.Set("Method", "DELETE")
		req.Method = "DELETE"
		req.Header.Set("X-Service-UUID", meta.UUID)
		req.Header.Set("X-Request-UUID", uuid.New())
		req.Header.Set("Content-Type", "charset=utf-8;application/json")

	}, func(resd ...interface{}) {
		cb(resd...)
	})

}

//Request provides a means of providing a generic requests to the server
func (hl *HTTPLink) Request(fpath, target string, body io.Reader, before func(r ...interface{}), after func(r ...interface{})) error {
	path := fmt.Sprintf("%s://%s/%s/%s", hl.GetDescriptor().Scheme, hl.GetPath(), hl.GetPrefix(), fpath)
	var req *http.Request
	var err error

	if body == nil {
		req, err = http.NewRequest("GET", path, body)

		if err != nil {
			return err
		}

	} else {
		req, err = http.NewRequest("POST", path, body)

		if err != nil {
			return err
		}
	}

	req.Header.Set("X-Service-Request", hl.GetPath())
	req.Header.Set("X-Service-Request-Target", target)

	before(req, target)

	res, err := hl.client.Do(req)

	if err != nil {
		log.Fatal("Response errored:", fpath, target, err)
		return err
	}

	defer res.Body.Close()

	bo, err := ioutil.ReadAll(res.Body)

	after(bo, res, req)

	return err

}
