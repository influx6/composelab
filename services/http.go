package services

import (
	"log"
	"net/http"

	"github.com/influx6/composelab/arch"
	"github.com/influx6/grids"
)

//HTTPService provides the service struct for all http services
type HTTPService struct {
	*arch.Service
}

//Dial beings the service connection
func (m *HTTPService) Dial() error {
	return http.ListenAndServe(m.GetPath(), http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		m.Route.IssueRequestPath(r.URL.Path, func(pack *grids.GridPacket) {
			pack.Set("Req", r)
			pack.Set("Res", rw)
			log.Println("sending off packet", r.URL.Path)
		})
	}))
}

//NewHTTPSlave creates a new slave service struct
func NewHTTPSlave(serviceName string, slaveAddr string, slavePort int, master arch.Linkage) *HTTPService {
	var sm = &HTTPService{arch.NewService(serviceName, slaveAddr, slavePort, master)}

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
