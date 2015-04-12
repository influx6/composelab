package links

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"

	"code.google.com/p/go-uuid/uuid"

	"github.com/influx6/composelab/arch"
	"github.com/influx6/goutils"
)

//UDPLink handles udp level communication
type UDPLink struct {
	*arch.ServiceLink
	Conn   *net.UDPConn
	ToAddr *net.UDPAddr
	MyAddr *net.UDPAddr
	buffer []byte
	closer chan interface{}
}

//NewUDPLink creates a new udp based service link
func NewUDPLink(serviceName string, addr string, port int) (*UDPLink, error) {
	address := fmt.Sprintf("%s:%d", addr, port)
	udpAddr, err := net.ResolveUDPAddr("udp4", address)

	if err != nil {
		log.Fatalf("Error generating udp address: %v %s %s %d", err, address, addr, port)
		return nil, err
	}

	cAddr, err := net.ResolveUDPAddr("udp4", ":0")

	if err != nil {
		log.Fatal("Error generating custom client udp address: ", err, cAddr)
		return nil, err
	}

	desc := arch.NewDescriptor("udp", serviceName, addr, udpAddr.Port, udpAddr.Zone, "udp4")

	return &UDPLink{
		arch.NewServiceLink(desc),
		nil,
		udpAddr,
		cAddr,
		make([]byte, 1024),
		make(chan interface{}),
	}, nil
}

//NewUDPWrap wraps a UDPLink as a arch.Linkage
func NewUDPWrap(h *UDPLink) arch.Linkage {
	return arch.Linkage(h)
}

//ReceiveDatagrams reads data from the server
func (u *UDPLink) ReceiveDatagrams() {
	for {
		select {
		case <-u.closer:
			u.Conn.Close()
		default:
			len, _, err := u.Conn.ReadFromUDP(u.buffer)

			if err != nil {
				log.Fatal("Error reading udp", err)
				// u.closer <- 1
				return
			}

			u.Send(u.buffer[:len])
		}
	}
}

//Dial startup the service link
func (u *UDPLink) Dial() {
	if u.Conn != nil {
		return
	}

	conn, err := net.DialUDP("udp", u.MyAddr, u.ToAddr)

	if err != nil {
		log.Fatalf("Error creating udp connection: %v %s", err, u.GetPath())
		return
	}

	u.Conn = conn
	//bootup and listen for data
	go u.ReceiveDatagrams()
}

//End calls to disconnect the udp link
func (u *UDPLink) End() {
	if u.closer == nil {
		return
	}

	u.closer <- 1
	close(u.closer)
	u.closer = nil
}

//Discover meets the Linkage interface to request discovery from a server
func (u *UDPLink) Discover(target string, callback func(string, interface{}, interface{})) error {
	return u.Request(fmt.Sprintf("%s/%s", "discover", target), target, nil, nil, func(d ...interface{}) {
		//do something interesting
		jsm := d[1]
		rsm := d[0]

		jsx, ok := rsm.(*arch.UDPPack)

		if !ok {
			log.Fatalf("response is not a UDPPack %v from %v ", rsm, jsm)
			return
		}

		var data interface{}

		err := json.Unmarshal(jsx.Data, data)

		if err != nil {
			smx := goutils.MorphString.Morph(jsx.Data)
			callback(target, smx, jsx)
			return
		}

		callback(target, data, jsx)

	})
}

//Unregister sends off a service meta information deregistration to the server
func (u *UDPLink) Unregister(target string, meta *arch.LinkDescriptor, callback func(data ...interface{})) error {
	jsm, err := json.Marshal(meta)
	// jsm, err := meta.MarshalJSON()

	if err != nil {
		log.Fatalf("error occured in coverting map with json %v %v", meta, err)
		return err
	}

	r, w := io.Pipe()

	go func() {
		defer w.Close()
		w.Write(jsm)
	}()

	err = u.Request(fmt.Sprintf("%s/%s", "unregister", target), target, r, nil, func(d ...interface{}) {
		//do something interesting
		if callback != nil {
			callback(d...)
		}
	})

	return err
}

//Register sends off a service meta information to the server
func (u *UDPLink) Register(target string, meta *arch.LinkDescriptor, callback func(data ...interface{})) error {
	jsm, err := json.Marshal(meta)
	// jsm, err := meta.MarshalJSON()

	if err != nil {
		log.Fatalf("error occured in coverting map with json %v %v", meta, err)
		return err
	}

	r, w := io.Pipe()

	go func() {
		defer w.Close()
		w.Write(jsm)
	}()

	err = u.Request(fmt.Sprintf("%s/%s", "register", target), target, r, nil, func(d ...interface{}) {
		//do something interesting
		if callback != nil {
			callback(d...)
		}
	})

	return err
}

//Request sends information to the server for a response
func (u *UDPLink) Request(tpath, target string, body io.Reader, before func(st ...interface{}), after func(smt ...interface{})) error {
	path := fmt.Sprintf("%s/%s", u.GetPrefix(), tpath)

	dataChan := make(chan []byte)

	go func() {
		defer close(dataChan)
		if body != nil {
			d, err := ioutil.ReadAll(body)

			if err == nil {
				dataChan <- d
			}

		} else {
			dataChan <- make([]byte, 0)
		}

	}()

	dat := <-dataChan

	jp := arch.NewUDPPack(path, target, uuid.New(), dat, u.MyAddr)

	if before != nil {
		before(jp, target)
	}

	// jsn, err := jp.MarshalJSON()
	jsn, err := json.Marshal(jp)

	if err != nil {
		return err
	}

	u.Conn.Write(jsn)

	ind := u.Size() + 1

	callable := func(data interface{}) {

		bo, ok := data.([]byte)

		if !ok {
			log.Fatal("Incorrect packet comming in", data, bo)
			return
		}

		jsnx := new(arch.UDPPack)
		err := json.Unmarshal(bo, jsnx)

		if err != nil {
			// strco := goutils.MorphString.Morph(bo)
			// if after != nil {
			after(bo, jp, target)
			// }
			return
		}

		if jp.UUID == jsnx.UUID {
			if after != nil {
				after(jsnx, jp, target)
			}
			u.ReceiveDoneOnce(func(d interface{}) {
				u.RemoveCallback(ind)
			})
		}

	}

	u.Receive(callable)

	return nil
}
