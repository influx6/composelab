package arch

import (
	"fmt"
	"io"
	"log"

	"github.com/influx6/composelab/routes"
	"github.com/influx6/evroll"
	"github.com/influx6/goutils"
	"github.com/influx6/grids"
)

var _ interface{}

//Services defines the interface that all services must implement to be considered valid servers
type Services interface {
	Dial() error
	Drop()
	Location() string
	Register(string, string, map[string]interface{})
	Unregister(string, string)
	Discover(string)
	Meta() map[string]interface{}
	ServiceName() string
}

//Linkage defines the link interface, links are like encapsulation of connection methods which allow the communication
//of services i.e your service can be communicating with udp to tcp(http) as protocols as far as it can be discovered and
//related to through http it is valid
type Linkage interface {
	GetPrefix() string
	GetPath() string
	GetAddress() string
	GetPort() int
	Discover(string, func(string, interface{})) error
	Register(string, map[string]interface{}) ([]byte, error)
	Request(string, io.Reader, func(...interface{})) ([]byte, error)
	Dial()
	End()
}

//Service is the base struct defining attributes of a service
type Service struct {
	*grids.Grid
	address     string
	port        int
	servicePath string
	Master      Linkage
	Slaves      *goutils.Map
	Registry    *goutils.Map
	Route       *routes.Routes
}

//ServiceLink is the concret struct define the linkage basic properties
type ServiceLink struct {
	*evroll.Streams
	prefix  string
	address string
	port    int
}

//Dial is an empty for handling service link dialing
func (s *ServiceLink) Dial() {
}

//End is an empty for handling service link disconnection
func (s *ServiceLink) End() {
}

//Discover is an empty for handling service link discover
func (s *ServiceLink) Discover(f string, b func(s string, f interface{})) error {
	return nil
}

//Request is an empty for handling service link discover
func (s *ServiceLink) Request(f string, bd io.Reader, action func(m ...interface{})) ([]byte, error) {
	return nil, nil
}

//Register is an empty for handling service link registeration for master operations
func (s *ServiceLink) Register(string, map[string]interface{}) ([]byte, error) {
	return nil, nil
}

//GetPrefix returns the prefix of the service
func (s *ServiceLink) GetPrefix() string {
	return s.prefix
}

//GetPath returns the path of the service
func (s *ServiceLink) GetPath() string {
	return fmt.Sprintf("%s:%d", s.GetAddress(), s.GetPort())
}

//GetAddress returns the address of the service
func (s *ServiceLink) GetAddress() string {
	return s.address
}

//GetPort returns the port of the service
func (s *ServiceLink) GetPort() int {
	return s.port
}

//NewServiceLink returns a new serviceLink
func NewServiceLink(prefix string, addr string, port int) *ServiceLink {
	return &ServiceLink{
		evroll.NewStream(false, false),
		prefix,
		addr,
		port,
	}
}

//NewService creates a new service struct
func NewService(serviceName string, addr string, port int, master Linkage) *Service {
	sv := &Service{
		grids.NewGrid(serviceName),
		addr,
		port,
		serviceName,
		master,
		goutils.NewMap(),
		goutils.NewMap(),
		routes.NewRoutes(serviceName),
	}

	sv.Route.Branch("discover")
	sv.Route.Branch("register")
	sv.Route.Branch("unregister")
	sv.Route.Branch("api")

	if sv.Master != nil {
		body, err := sv.Master.Register(serviceName, sv.Meta())

		if err != nil {
			log.Fatal("unable to register with master:", serviceName, body, err)
		}

	}

	return sv
}

//Divert provides a shortcut member funcs to call Divert on the Service Route
func (s *Service) Divert(sm *Service) {
	_ = s.Route.Divert(sm.Route)
}

//Branch provides a shortcut member funcs to call Branch on the Service Route
func (s *Service) Branch(path string) {
	s.Route.Branch(path)
}

//Select provides a shortcut member funcs to call Select on the Service Route
func (s *Service) Select(path string) (*routes.Routes, error) {
	return s.Route.Select(path)
}

//GetPath returns the path of the service
func (s *Service) GetPath() string {
	return fmt.Sprintf("%s:%d", s.address, s.port)
}

//Meta returns a map containing details about the service
func (s *Service) Meta() map[string]interface{} {
	return map[string]interface{}{
		"address": s.address,
		"port":    s.port,
		"service": s.ServiceName(),
	}
}

//Dial initiates the connections for the service
func (s *Service) Dial() error { return nil }

//ServiceName returns the service name/id
func (s *Service) ServiceName() string {
	return s.servicePath
}

//Drop stops and ends the connection for the service
func (s *Service) Drop() {}

//Location returns a string of the address and path of the service
func (s *Service) Location() string {
	return fmt.Sprintf("%s@%s", s.ServiceName(), s.GetPath())
}

//Register adds a servicelink into the services connection pool
func (s *Service) Register(serviceName string, uuid string, meta map[string]interface{}) {
	if s.Registry.Has(serviceName) {

		sector, ok := s.Registry.Get(serviceName).(*goutils.Map)

		if !ok {
			return
		}

		if sector.Has(uuid) {
			return
		}

		sector.Set(uuid, meta)
	} else {
		smap := goutils.NewMap()
		s.Registry.Set(serviceName, smap)
		smap.Set(uuid, meta)
	}

}

//Unregister removes a servicelink from the services connection pool
func (s *Service) Unregister(serviceName string, uuid string) {
	if !s.Registry.Has(serviceName) {
		return
	}

	sector, ok := s.Registry.Get(serviceName).(*goutils.Map)

	if !ok {
		return
	}

	if !sector.Has(uuid) {
		return
	}

	sector.Remove(uuid)
}

//Discover is an empty for handling service link discover
func (s *Service) Discover(string) error {
	return nil
}

//Request is an empty for handling service link discover
func (s *Service) Request(string) {

}
