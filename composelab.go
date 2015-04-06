package composelab

import "github.com/influx6/composelab/services"

//Master struct for master connections
type Master struct {
	*services.HTTPService
}

//NewMaster creates a new master service struct
func NewMaster(addr string, port int) *Master {
	sm := services.NewHTTPSlave("master", addr, port, nil)

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
