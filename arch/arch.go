package arch

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"

	"code.google.com/p/go-uuid/uuid"
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
	Discover(string)
	ServiceName() string
	Register(string, string, *LinkDescriptor)
	Unregister(string)
	HasRegistered(string) bool
	GetServices(string) *LinkDescriptor
}

//Linkage defines the link interface, links are like encapsulation of connection methods which allow the communication
//of services i.e your service can be communicating with udp to tcp(http) as protocols as far as it can be discovered and
//related to through http it is valid
type Linkage interface {
	GetDescriptor() *LinkDescriptor
	GetPrefix() string
	GetPath() string
	GetAddress() string
	GetPort() int
	Discover(string, func(string, interface{}, interface{})) error
	Register(string, *LinkDescriptor, func(...interface{})) error
	Request(string, string, io.Reader, func(...interface{}), func(...interface{})) error
	Dial()
	End()
}

//UDPPack represents a standard json udp message
type UDPPack struct {
	Path    string       `json:"path"`
	Service string       `json:"service"`
	UUID    string       `json:"uuid"`
	Data    []byte       `json:"data"`
	Address *net.UDPAddr `json:"address"`
	// Visited []*net.UDPAddr `json:"visited"`
}

//NewUDPPack creates a new udp packet
func NewUDPPack(path, service, uuid string, data []byte, addr *net.UDPAddr) *UDPPack {
	return &UDPPack{
		path,
		service,
		uuid,
		data,
		addr,
	}
}

//UDPPackFrom creates a new udp packet from a previous one with only the data
//and addr changed
func UDPPackFrom(u *UDPPack, data []byte, addr *net.UDPAddr) *UDPPack {
	return NewUDPPack(u.Path, u.Service, u.UUID, data, addr)
}

//MarshalJSON returns the json byte version of the LinkDescriptor
// func (u *UDPPack) MarshalJSON() ([]byte, error) {
// 	lb, err := json.Marshal(u)
// 	return lb, err
// }

//Service is the base struct defining attributes of a service
type Service struct {
	*grids.Grid
	descrptior *LinkDescriptor
	Master     Linkage
	Slaves     *goutils.Map
	registry   *goutils.Map
	Route      *routes.Routes
}

//LinkDescriptor provides basic level description for links
type LinkDescriptor struct {
	Service string                 `json:"service"`
	Address string                 `json:"address"`
	Port    int                    `json:"port"`
	Zone    string                 `json:"zone"`
	Scheme  string                 `json:"scheme"`
	Misc    map[string]interface{} `json:"misc"`
	Proto   string                 `json:"proto"`
	UUID    string                 `json:"uuid"`
}

// MarshalJSON returns the json byte version of the LinkDescriptor
func (l *LinkDescriptor) MarshalJSON() ([]byte, error) {
	lb, err := json.Marshal(l)
	return lb, err
}

//NewDescriptor creates a new LinkDescriptor
func NewDescriptor(proto string, name string, addr string, port int, zone string, scheme string) *LinkDescriptor {
	return &LinkDescriptor{
		name,
		addr,
		port,
		zone,
		scheme,
		make(map[string]interface{}),
		proto,
		uuid.New(),
	}
}

//ServiceLink is the concret struct define the linkage basic properties
type ServiceLink struct {
	*evroll.Streams
	desc *LinkDescriptor
}

//GetDescriptor is an empty for handling service link dialing
func (s *ServiceLink) GetDescriptor() *LinkDescriptor {
	return s.desc
}

//Dial is an empty for handling service link dialing
func (s *ServiceLink) Dial() {
}

//End is an empty for handling service link disconnection
func (s *ServiceLink) End() {
}

//Discover is an empty for handling service link discover
func (s *ServiceLink) Discover(f string, b func(s string, data interface{}, res interface{})) error {
	return nil
}

//Request is an empty for handling service link discover
func (s *ServiceLink) Request(f string, target string, bd io.Reader, before func(m ...interface{}), after func(m ...interface{})) error {
	return nil
}

//Register is an empty for handling service link registeration for master operations
func (s *ServiceLink) Register(sm string, m *LinkDescriptor, cf func(sets ...interface{})) error {
	return nil
}

//GetUUID returns the UUID string of the service link
func (s *ServiceLink) GetUUID() string {
	return s.desc.UUID
}

//GetPrefix returns the prefix of the service link
func (s *ServiceLink) GetPrefix() string {
	return s.desc.Service
}

//GetPath returns the path of the service link
func (s *ServiceLink) GetPath() string {
	return fmt.Sprintf("%s:%d", s.GetAddress(), s.GetPort())
}

//GetAddress returns the address of the service
func (s *ServiceLink) GetAddress() string {
	return s.desc.Address
}

//GetPort returns the port of the service
func (s *ServiceLink) GetPort() int {
	return s.desc.Port
}

//NewServiceLink returns a new serviceLink
func NewServiceLink(d *LinkDescriptor) *ServiceLink {
	return &ServiceLink{
		evroll.NewStream(false, false),
		d,
	}
}

//NewService creates a new service struct
func NewService(desc *LinkDescriptor, master Linkage) *Service {
	sv := &Service{
		grids.NewGrid(desc.Service),
		desc,
		master,
		goutils.NewMap(),
		goutils.NewMap(),
		routes.NewRoutes(desc.Service),
	}

	sv.Route.Branch("discover")
	sv.Route.Branch("register")
	sv.Route.Branch("unregister")
	sv.Route.Branch("api")

	if sv.Master != nil {
		err := sv.Master.Register(desc.Service, sv.GetDescriptor(), func(d ...interface{}) {
			//maybe
		})

		if err != nil {
			log.Fatal("unable to register with master:", desc.Service, err)
		}

	}

	return sv
}

//GetDescriptor is an empty for handling service link dialing
func (s *Service) GetDescriptor() *LinkDescriptor {
	return s.descrptior
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
	return fmt.Sprintf("%s:%d", s.GetAddress(), s.GetPort())
}

//GetAddress returns the adddress of the service
func (s *Service) GetAddress() string {
	return s.descrptior.Address
}

//GetPort returns the adddress of the service
func (s *Service) GetPort() int {
	return s.descrptior.Port
}

//Dial initiates the connections for the service
func (s *Service) Dial() error { return nil }

//ServiceName returns the service name/id
func (s *Service) ServiceName() string {
	return s.descrptior.Service
}

//Drop stops and ends the connection for the service
func (s *Service) Drop() {}

//Location returns a string of the address and path of the service
func (s *Service) Location() string {
	return fmt.Sprintf("%s@%s", s.ServiceName(), s.GetPath())
}

//Register adds a servicelink into the services connection pool
func (s *Service) Register(serviceName string, meta *LinkDescriptor) {
	if !s.registry.Has(serviceName) {
		s.registry.Set(serviceName, meta)
	}
}

//Unregister removes a servicelink from the services connection pool
func (s *Service) Unregister(serviceName string) {
	s.registry.Remove(serviceName)
}

//HasRegistered checks whether a particular service of a specific serviceName is registered
//and if supplied checks whether there exists a provider with the uuid
func (s *Service) HasRegistered(serviceName string) bool {
	return s.registry.Has(serviceName)
}

//GetServiceProvider returns a goutils.Map containing register services under the
//serviceName provided and supplies a secondary error argument to indicate
//error
func (s *Service) GetServiceProvider(serviceName string) (*LinkDescriptor, error) {
	if !s.registry.Has(serviceName) {
		return nil, fmt.Errorf("%s not found", serviceName)
	}

	li, ok := s.registry.Get(serviceName).(*LinkDescriptor)

	if !ok {
		return nil, fmt.Errorf("%s unable to convert to LinkDescriptor %v", serviceName, li)
	}

	return li, nil
}
