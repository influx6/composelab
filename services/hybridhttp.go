package sevices

import (
	"net/http"
	"net/url"

	"github.com/influx6/composelab/arch"
)

//JSONPService represents a service handling the jsonp protocol
type JSONPService struct {
	*HTTPService
	callback string
}

//WebSocketService represents a service handling the websocket protocol
type WebSocketService struct {
	*HTTPService
}

//PollService represents a service handling the http-long polling protocol
type PollService struct {
	*HTTPService
}

//ProcessPackets for JSONPService handles checking and validating a request as a
//valid jsonp request before process it for use
func (j *JSONPService) ProcessPackets(rw http.ResponseWriter, r *http.Request) {
	q, _ := url.ParseQuery(r.URL.RawQuery)

	if _, ok := q["json"]; ok {
		if _, ok := q[j.callback]; ok {
			j.HTTPService.processPackets(rw, r)
		}
	}
}

//NewJSONPService returns a new jsonp based service struct
func NewJSONPService(service, addr string, port int, master arch.Linkage, callbackName sring) {
	hs := NewHTTPService(service, addr, port, master)
	var cb string

	if callbackName == "" {
		cb = "Callback"
	} else {
		cb = callbackName
	}

	return JSONPService{
		hs,
		cb,
	}
}
