package links

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/influx6/composelab/arch"
)

//HTTPLink represents a standard servicelink using http underneath
//as its protocol transport
type HTTPLink struct {
	*arch.ServiceLink
	client *http.Client
	scheme string
}

//NewHTTPLink returns a new http service link
func NewHTTPLink(prefix string, addr string, port int) *HTTPLink {
	return &HTTPLink{
		arch.NewServiceLink(prefix, addr, port),
		new(http.Client),
		"http://",
	}
}

//NewSecureHTTPLink returns a new http service link
func NewSecureHTTPLink(prefix string, addr string, port int, trans *http.Transport) *HTTPLink {
	cl := &http.Client{Transport: trans}
	return &HTTPLink{
		arch.NewServiceLink(prefix, addr, port),
		cl,
		"https://",
	}
}

//NewHTTPWrap wraps a HTTPLink as a arch.Linkage
func NewHTTPWrap(h *HTTPLink) arch.Linkage {
	return arch.Linkage(h)
}

//Discover sends a request to the set server links if a service exists
func (hl *HTTPLink) Discover(target string, callback func(string, interface{})) error {
	path := []string{hl.scheme, hl.GetPrefix(), "discover", target}
	res, err := hl.client.Get(strings.Join(path, "/"))

	if err != nil {
		log.Fatal(err)
		return err
	}

	defer res.Body.Close()
	status := res.StatusCode

	if status == 200 {
		body, err := ioutil.ReadAll(res.Body)

		if err != nil {
			log.Fatal(err)
			return err
		}

		jsn := map[string]interface{}{}

		err = json.Unmarshal(body, jsn)

		if err != nil {
			log.Fatal(err)
			return err
		}

		callback(target, jsn)
	}

	return nil
}

//Register  registers a service to the specific server with the meta details as json
func (hl *HTTPLink) Register(target string, meta map[string]interface{}) ([]byte, error) {
	path := []string{"register", target}
	jsn, err := json.Marshal(meta)
	url := strings.Join(path, "/")

	if err != nil {
		return nil, err
	}

	res, err := hl.Request(url, bytes.NewReader(jsn), func(sets ...interface{}) {
		rq := sets[0]
		req, ok := rq.(*http.Request)

		if !ok {
			return
		}

		req.Header.Set("X-Service-Request", hl.GetPath())
		req.Header.Set("Content-Type", "application/json")
	})

	if err != nil {
		return nil, err
	}

	return res, nil

}

//Request provides a means of providing a generic requests to the server
func (hl *HTTPLink) Request(target string, body io.Reader, handler func(r ...interface{})) ([]byte, error) {
	path := []string{hl.scheme, hl.GetPrefix(), target}
	var req *http.Request
	var err error

	if body == nil {
		req, err = http.NewRequest("GET", strings.Join(path, "/"), body)

		if err != nil {
			return nil, err
		}

	} else {
		req, err = http.NewRequest("POST", strings.Join(path, "/"), body)

		if err != nil {
			return nil, err
		}
	}

	handler(req)

	res, err := hl.client.Do(req)

	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	bo, err := ioutil.ReadAll(res.Body)

	return bo, err

}
