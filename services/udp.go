package services

import (
	"encoding/json"
	"fmt"
	"log"
	"net"

	"github.com/influx6/composelab/arch"
	"github.com/influx6/grids"
)

//UDPService provides the service struct for all udp services
type UDPService struct {
	*arch.Service
	closer chan interface{}
	buffer []byte
	Addr   *net.UDPAddr
	Server *net.UDPConn
}

//ProcessDatagrams handles the process of datagrams
func (u *UDPService) ProcessDatagrams() {
	for {
		select {
		case <-u.closer:
			log.Fatal("UDPServer is shutting down", u)
			u.Server.Close()
		default:
			len, _, err := u.Server.ReadFromUDP(u.buffer)

			if err != nil {
				log.Fatal("Error reading udp", err)
				return
			}

			data := u.buffer[:len]

			upack := new(arch.UDPPack)
			err = json.Unmarshal(data, upack)

			if err != nil {
				log.Fatal("data is not a packet", err)
				return
			}

			u.Route.IssueRequestPath(upack.Path, func(p *grids.GridPacket) {
				p.Set("Packet", upack)
			})

		}
	}
}

//Dial sets up the server and connections for the server
func (u *UDPService) Dial() {
	if u.Server != nil {
		return
	}

	con, err := net.ListenUDP("udp", u.Addr)

	if err != nil {
		log.Fatal("UDP Server dialing error", err, con, u.Addr)
		return
	}

	u.Server = con
	go u.ProcessDatagrams()
}

//End calls to disconnect the udp link
func (u *UDPService) End() {
	if u.Server == nil {
		return
	}

	u.closer <- 1
	close(u.closer)
	u.Server = nil
}

//NewUDPService returns a new udp service struct
func NewUDPService(serviceName string, addr string, port int, master arch.Linkage) (*UDPService, error) {
	uaddr, err := net.ResolveUDPAddr("udp4", fmt.Sprintf("%s:%d", addr, port))

	if err != nil {
		log.Fatal("Error occured creating udp address", uaddr, port, addr)
		return nil, err
	}

	var um = &UDPService{
		arch.NewService(serviceName, addr, port, master),
		make(chan interface{}),
		make([]byte, 1024),
		uaddr,
		nil,
	}

	reg, err := um.Select("register")

	if err == nil {
		reg.Terminal().Only(grids.ByPackets(func(g *grids.GridPacket) {
			log.Println("/register receieves", g)
		}))
	}

	disc, err := um.Select("discover")

	if err == nil {
		disc.Terminal().Only(grids.ByPackets(func(g *grids.GridPacket) {
			log.Println("/discover receieves", g)
		}))
	}

	unreg, err := um.Select("unregister")

	if err == nil {
		unreg.Terminal().Only(grids.ByPackets(func(g *grids.GridPacket) {
			log.Println("/unregister receieves", g)
		}))
	}

	return um, nil
}
