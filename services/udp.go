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
			len, addr, err := u.Server.ReadFromUDP(u.buffer)

			if err != nil {
				log.Fatal("Error reading udp", err, addr)
				return
			}

			data := u.buffer[:len]

			upack := new(arch.UDPPack)
			err = json.Unmarshal(data, upack)

			if err != nil {
				log.Fatal("data is not a valid udp service packet", err)
				return
			}

			upack.Address = addr

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
	u.ProcessDatagrams()
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

//UDPNorm type that specifies type descriptor for WhenUDP
type UDPNorm func(*arch.LinkDescriptor, *arch.UDPPack)

//WhenUDP provides a callback adaptor to check and collect udp data from a gridPacket
var WhenUDP = func(checkDesc bool, g *grids.GridPacket, norm UDPNorm) {
	if !g.Has("Packet") {
		return
	}

	udp, ok := g.Get("Packet").(*arch.UDPPack)

	if !ok {
		return
	}

	if checkDesc {
		dc := new(arch.LinkDescriptor)

		err := json.Unmarshal(udp.Data, dc)

		if err != nil {
			log.Fatal("Error occur while parsing json data buffer ", err, udp.Data, udp)
			return
		}

		norm(dc, udp)
		return
	}

	norm(nil, udp)
}

//ResponseError responds to a udp pack with a generic error map
var ResponseError = func(u *arch.UDPPack, um *UDPService) {
	ub := arch.UDPPackFrom(u, []byte("{ error:'not found!'}"), um.Addr)
	ubinx, err := json.Marshal(ub)

	if err != nil {
		log.Fatal("Unable to create udppack for: ", err, ubinx, ub)
		return
	}

	um.Server.WriteTo(ubinx, u.Address)
}

//ResponseSuccess response to a udp pack with a generic success map
var ResponseSuccess = func(u *arch.UDPPack, um *UDPService) {
	ub := arch.UDPPackFrom(u, []byte("{ state: 200, error: 'nil'}"), um.Addr)
	ubinx, err := json.Marshal(ub)

	if err != nil {
		log.Fatal("Unable to create udppack for: ", err, ubinx, ub)
		return
	}

	um.Server.WriteTo(ubinx, u.Address)
}

//NewUDPService returns a new udp service struct
func NewUDPService(serviceName string, addr string, port int, master arch.Linkage) (*UDPService, error) {
	uaddr, err := net.ResolveUDPAddr("udp4", fmt.Sprintf("%s:%d", addr, port))

	if err != nil {
		log.Fatal("Error occured creating udp address", uaddr, port, addr)
		return nil, err
	}

	desc := arch.NewDescriptor("udp", serviceName, addr, uaddr.Port, uaddr.Zone, "udp4")

	var um = &UDPService{
		arch.NewService(desc, master),
		make(chan interface{}),
		make([]byte, 1024),
		uaddr,
		nil,
	}

	reg, err := um.Select("register")

	if err == nil {
		reg.Terminal().Any(grids.ByPackets(func(g *grids.GridPacket) {
			WhenUDP(true, g, func(li *arch.LinkDescriptor, u *arch.UDPPack) {
				um.Register(u.Service, li)
				ResponseSuccess(u, um)
			})
		}))
	}

	disc, err := um.Select("discover")

	if err == nil {
		disc.Terminal().Any(grids.ByPackets(func(g *grids.GridPacket) {
			log.Println("/discover receieves", g)
			WhenUDP(false, g, func(_ *arch.LinkDescriptor, u *arch.UDPPack) {
				if um.HasRegistered(u.Service) {
					li, err := um.GetServiceProvider(u.Service)

					if err != nil {
						log.Fatal("Unable to find service: ", u.Service, u)
						return
					}

					bin, err := json.Marshal(li)
					// bin, err := li.MarshalJSON()

					if err != nil {
						log.Fatal("Unable to jsonify service linkdescriptor: ", err, li)
						return
					}

					ub := arch.UDPPackFrom(u, bin, um.Addr)
					ubinx, err := json.Marshal(ub)

					if err != nil {
						log.Fatal("Unable to create udppack for: ", err, li, ubinx, ub)
						return
					}

					um.Server.WriteTo(ubinx, u.Address)

				} else {
					log.Println("responding with error")
					ResponseError(u, um)
				}
			})
		}))
	}

	unreg, err := um.Select("unregister")

	if err == nil {
		unreg.Terminal().Any(grids.ByPackets(func(g *grids.GridPacket) {
			log.Println("/unregister receieves", g)
			WhenUDP(true, g, func(li *arch.LinkDescriptor, u *arch.UDPPack) {
				um.Unregister(u.Service, li)
				ResponseSuccess(u, um)
			})
		}))
	}

	return um, nil
}
