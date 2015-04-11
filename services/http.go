package services

import (
	"log"
	"net/http"

	"github.com/influx6/composelab/arch"
	"github.com/influx6/grids"
)

//HTTPCert provides a simple struct to hold http tls certificate and key
type HTTPCert struct {
	Key  string
	Cert string
}

//HTTPService provides the service struct for all http services
type HTTPService struct {
	*arch.Service
	cert *HTTPCert
}

//Dial beings the service connection
func (m *HTTPService) Dial() error {
	if m.cert != nil {
		return http.ListenAndServe(m.GetPath(), http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			m.Route.IssueRequestPath(r.URL.Path, func(pack *grids.GridPacket) {
				pack.Set("Req", r)
				pack.Set("Res", rw)
				log.Println("sending off packet", r.URL.Path)
			})
		}))
	} else {
		return http.ListenAndServe(m.GetPath(), http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			m.Route.IssueRequestPath(r.URL.Path, func(pack *grids.GridPacket) {
				pack.Set("Req", r)
				pack.Set("Res", rw)
				log.Println("sending off packet", r.URL.Path)
			})
		}))
	}
}

//NewHTTPFactory creates a new slave service struct
func NewHTTPFactory(serviceName string, slaveAddr string, slavePort int, cert *HTTPCert, master arch.Linkage) *HTTPService {
	var scheme string

	if cert != nil {
		scheme = "http://"
	} else {
		scheme = "https://"
	}

	desc := arch.NewDescriptor("http", serviceName, slaveAddr, slavePort, "0", scheme)
	var sm = &HTTPService{arch.NewService(desc, master), cert}

	reg, err := sm.Select("register")

	if err == nil {
		reg.Terminal().Only(grids.ByPackets(func(g *grids.GridPacket) {
			log.Println("/register receieves", g)
		}))
	}

	disc, err := sm.Select("discover")

	if err == nil {
		disc.Terminal().Only(grids.ByPackets(func(g *grids.GridPacket) {
			log.Println("/discover receieves", g)
		}))
	}

	unreg, err := sm.Select("unregister")

	if err == nil {
		unreg.Terminal().Only(grids.ByPackets(func(g *grids.GridPacket) {
			log.Println("/unregister receieves", g)
		}))
	}

	return sm
}

//NewHTTPService creates a new http service struct
func NewHTTPService(service, addr string, port int, master arch.Linkage) *HTTPService {
	return NewHTTPFactory(service, addr, port, nil, master)
}

//NewHTTPSecureService creates a new secure http service struct
func NewHTTPSecureService(service, addr string, port int, cert *HTTPCert, master arch.Linkage) *HTTPService {
	return NewHTTPFactory(service, addr, port, cert, master)
}
