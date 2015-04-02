package arch

import (
	"fmt"
	"strings"

	"github.com/influx6/composelab/routes"
	"github.com/influx6/goutils"
)

//Services defines the interface that all services must implement to be considered valid servers
type Services interface {
	Dial()
	Drop()
	Location() string
	Register(string, string, map[string]interface{})
	Unregister(string, string)
	Discover(string)
}

//Linkage defines the link interface, links are like encapsulation of connection methods which allow the communication
//of services i.e your service can be communicating with udp to tcp(http) as protocols as far as it can be discovered and
//related to through http it is valid
type Linkage interface {
	GetPath() string
	GetAddress() string
	GetPort() int
	Serve()
	Dial()
	Drop()
}

//Service is the base struct defining attributes of a service
type Service struct {
	ServicePath string
	Master      Linkage
	Hook        Linkage
	Slaves      *goutils.Map
	Routes      *routes.Routes
}

//Master struct for master connections
type Master struct {
	*Service
}

//ServiceLink is the concret struct define the linkage basic properties
type ServiceLink struct {
	Address string
	Port    int
}

//NewServiceLink returns a new serviceLink
func NewServiceLink(addr string, port int) *ServiceLink {
	return &ServiceLink{
		addr,
		port,
	}
}

//WrapLink returns a serviceLink wrapped by a Linkage
func WrapLink(sl *ServiceLink) Linkage {
	return Linkage(sl)
}

//HttpLink retuns a new service link for http based operations
type HttpLink struct {
	*ServiceLink
}

//NewHttpLink returns a new http service link
func NewHttpLink(addr string, port int) *HttpLink {
	return &HttpLink{NewServiceLink(addr, port)}
}

//NewService creates a new service struct
func NewService(serviceName string, master Linkage, hook Linkage) *Service {
	sv := &Service{
		serviceName,
		master,
		hook,
		goutils.NewMap(),
		routes.NewRoutes(serviceName),
	}

	sv.Routes.Branch("discover")
	sv.Routes.Branch("register")
	sv.Routes.Branch("unregister")
	sv.Routes.Branch("api")

	return sv
}

//NewMaster creates a new master service struct
func NewMaster(addr string, port int) *Service {
	hook := Linkage(NewHttpLink(addr, port))
	return NewService("master", nil, hook)
}

//NewHttpSlave creates a new slave service struct
func NewHttpSlave(serviceName string, slaveAddr string, masterAddr string) (*Service, error) {
	sa := strings.Split(slaveAddr, ":")
	ma := strings.Split(masterAddr, ":")

	if len(sa) < 2 {
		return nil, fmt.Errorf("Slave address incorrect, expecting 'addr:port' format %s", slaveAddr)
	}

	if len(ma) < 2 {
		return nil, fmt.Errorf("Master address incorrect, expecting 'addr:port' format %s", masterAddr)
	}

	sad := sa[0]
	spt := int(sa[1])

	mad := ma[0]
	mpt := int(ma[1])

	master := Linkage(NewHttpLink(mad, mpt))
	hook := Linkage(NewHttpLink(sad, spt))
	return NewService(serviceName, master, hook), nil
}

//Dial initiates the connections for the service
func (s *Service) Dial() {
	s.Hook.Dial()
}

//Drop stops and ends the connection for the service
func (s *Service) Drop() {
	s.Hook.Drop()
}

//Location returns a string of the address and path of the service
func (s *Service) Location() string {
	return fmt.Sprintf("%s@%s", s.ServicePath, s.Hook.GetPath())
}

//GetPath returns the path of the service
func (s *ServiceLink) GetPath() string {
	return fmt.Sprintf("%s:%d", s.GetAddress(), s.GetPort())
}

//GetAddress returns the address of the service
func (s *ServiceLink) GetAddress() string {
	return s.Address
}

//GetPort returns the port of the service
func (s *ServiceLink) GetPort() int {
	return s.Port
}

//Dial provides the connection features for a servicelink
func (s *ServiceLink) Dial() {

}

//Drop provides the disconnection features for a servicelink
func (s *ServiceLink) Drop() {
}

//Serve provides the requests features for a servicelink
func (s *ServiceLink) Serve() {
}

//Register adds a servicelink into the services connection pool
func (s *Service) Register(serviceName string, uuid string, meta map[string]interface{}) {
	if s.Slaves.Has(serviceName) {

		sector, ok := s.Slaves.Get(serviceName).(*goutils.Map)

		if !ok {
			return
		}

		if sector.Has(uuid) {
			return
		}

		sector.Set(uuid, meta)
	}
}

//Unregister removes a servicelink from the services connection pool
func (s *Service) Unregister(serviceName string, uuid string) {
	if !s.Slaves.Has(serviceName) {
		return
	}

	sector, ok := s.Slaves.Get(serviceName).(*goutils.Map)

	if !ok {
		return
	}

	if !sector.Has(uuid) {
		return
	}

	// suid := sector.Get(uuid).(Linkage)
	sector.Remove(uuid)
	// suid.Drop()
}

//Discover sends a request to the connected master seeking information about
//about a service if such exists
func (s *Service) Discover(serviceName string) {

}
