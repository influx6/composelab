package sevices

import (
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/influx6/composelab/arch"
	"github.com/influx6/grids"
)

var webSocketUpgrade = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

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

func isWebSocketRequest(r *http.Request) bool {
	var _ interface{}
	_, hasupgrade := r.Header["Upgrade"]
	_, hassec := r.Header["Sec-Websocket-Version"]
	_, hasext := r.Header["Sec-Websocket-Extensions"]
	_, haskey := r.Header["Sec-Websocket-Key"]
	return hasupgrade && hassec && hasext && haskey
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

//ProcessPackets for JSONPService handles checking and validating a request as a
//valid jsonp request before process it for use
func (w *WebSocketService) ProcessPackets(rw http.ResponseWriter, r *http.Request) {
	q, _ := url.ParseQuery(r.URL.RawQuery)

	if isWebSocketRequest(r) {
		rw.Header().Set("Upgrades", strings.Join(proc, ";"))

		agent, ok := r.Header["User-Agent"]

		if ok {
			ag := strings.Join(agent, ";")
			msie := strings.Index(ag, ";MSIE")
			trident := strings.Index(ag, "Trident/")

			if msie != -1 || trident != -1 {
				rw.Header().Set("X-XSS-Protection", "0")
			}
		}

		origin, ok := r.Header["Origin"]

		if ok {
			rw.Header().Set("Access-Control-Allow-Credentials", "true")
			rw.Header().Set("Access-Control-Allow-Origin", strings.Join(origin, ";"))
		} else {
			rw.Header().Set("Access-Control-Allow-Origin", "*")
		}

		u, err := webSocketUpgrade.Upgrade(rw, r, nil)

		if err != nil {
			log.Println(err)
			return
		}

		w.Route.IssueRequestPath(r.URL.Path, func(pack *grids.GridPacket) {
			pack.Set("Ws", u)
			pack.Set("Req", r)
			pack.Set("Res", rw)
			CollectHTTPBody(r, pack)
		})
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
