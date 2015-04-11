package composelab

import "github.com/influx6/composelab/services"

//Master struct for master connections
type Master struct {
	*services.HTTPService
}

//NewMaster creates a new master service struct
func NewMaster(addr string, port int, cert *services.HTTPCert) *Master {
	var sm *services.HTTPService

	if cert == nil {
		sm = services.NewHTTPService("master", addr, port, nil)
	} else {
		sm = services.NewHTTPSecureService("master", addr, port, cert, nil)
	}

	// reg, err := sm.Select("register")
	//
	// if err == nil {
	//
	// 	reg.Terminal().Only(grids.ByPackets(func(g *grids.GridPacket) {
	// 		fmt.Println("/register receieves", g)
	// 	}))
	// }
	//
	// disc, err := sm.Select("discover")
	//
	// if err == nil {
	//
	// 	disc.Terminal().Only(grids.ByPackets(func(g *grids.GridPacket) {
	// 		fmt.Println("/discover receieves", g)
	// 	}))
	// }
	//
	// unreg, err := sm.Select("unregister")
	//
	// if err == nil {
	//
	// 	unreg.Terminal().Only(grids.ByPackets(func(g *grids.GridPacket) {
	// 		fmt.Println("/unregister receieves", g)
	// 	}))
	// }

	return &Master{sm}
}
