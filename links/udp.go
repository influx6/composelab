package links

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"

	"code.google.com/p/go-uuid/uuid"
	"github.com/influx6/composelab/arch"
)

//UDPLink handles udp level communication
type UDPLink struct {
	*arch.ServiceLink
	Conn *net.UDPConn
	Addr *net.UDPAddr
}

//UDPPack represents a standard json udp message
type UDPPack struct {
	Path    string `json:"path"`
	Service string `json:"service"`
	UUID    string `json:"uuid"`
	// Meta    map[string]interface{} `json:"meta"`
	Data interface{} `json:"data"`
}

//NewUDPLink creates a new udp based service link
func NewUDPLink(serviceName string, addr string, port int) (*UDPLink, error) {
	address := fmt.Sprintf("%s:%d", addr, port)
	udpAddr, err := net.ResolveUDPAddr("udp4", address)
	cAddr, err := net.ResolveUDPAddr("udp4", ":0")

	if err != nil {
		log.Fatalf("Error generating udp address: %v %s %s %d", err, address, addr, port)
		return nil, err
	}

	conn, err := net.DialUDP("udp", cAddr, udpAddr)

	if err != nil {
		log.Fatalf("Error creating udp connection: %v %s", err, address)
		return nil, err
	}

	return &UDPLink{
		arch.NewServiceLink(serviceName, addr, port),
		conn,
		udpAddr,
	}, nil
}

//NewUDPWrap wraps a UDPLink as a arch.Linkage
func NewUDPWrap(h *UDPLink) arch.Linkage {
	return arch.Linkage(h)
}

//Dial startup the service link
func (u *UDPLink) Dial() {

}

//Discover meets the Linkage interface to request discovery from a server
func (u *UDPLink) Discover(target string, callback func(string, interface{})) error {

	return nil
}

//Register sends off a service meta information to the server
func (u *UDPLink) Register(target string, meta map[string]interface{}) ([]byte, error) {

	return nil, nil
}

//Request sends information to the server for a response
func (u *UDPLink) Request(target string, body io.Reader, callback func(st ...interface{})) ([]byte, error) {
	jp := &UDPPack{
		fmt.Sprintf("%s/%s", u.GetPrefix(), target),
		target,
		uuid.New(),
		nil,
	}

	jsn, err := json.Marshal(jp)

	if err != nil {
		return nil, err
	}

	u.Conn.Write(jsn)

	ind := u.Size() + 1

	callable := func(data interface{}) {

		reply, ok := data.([]byte)

		if !ok {
			return
		}

		var jsnx interface{}
		err := json.Unmarshal(reply, jsnx)

		if err != nil {
			return
		}

		mp, ok := jsnx.(UDPPack)

		if !ok {
			return
		}

		if jp.UUID == mp.UUID {
			callback(jp, mp, target)
			u.ReceiveDoneOnce(func(d interface{}) {
				u.RemoveCallback(ind)
			})
		}

	}

	u.Receive(callable)

	return nil, nil
}
