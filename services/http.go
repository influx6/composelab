package services

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

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

//CollectHTTPBody takes a requests and retrieves the body from the into a gridpacket object
func CollectHTTPBody(r *http.Request, g *grids.GridPacket) {
	content, ok := r.Header["Content-Type"]
	muxcontent := strings.Join(content, ";")
	wind := strings.Index(muxcontent, "application/x-www-form-urlencode")
	mind := strings.Index(muxcontent, "multipart/form-data")
	jsn := strings.Index(muxcontent, "application/json")

	if ok {

		if jsn != -1 {
			g.Set("Type", "json")
			g.Set("JSON", true)
			g.Set("Data", true)
		}

		if wind != -1 {
			if err := r.ParseForm(); err != nil {
				log.Println("Request Read Form Error", err)
			} else {
				form := r.Form
				pform := r.PostForm

				g.Set("Data", true)
				g.Set("Form", true)
				g.Set("Form", form)
				g.Set("PostForm", pform)
			}
		}

		if mind != -1 {
			if err := r.ParseMultipartForm(32 << 20); err != nil {
				log.Println("Request Read MultipartForm Error", err)
			} else {
				g.Set("Data", true)
				g.Set("Form", false)
				g.Set("Value", r.MultipartForm.Value)
				g.Set("Body", r.MultipartForm.File)
			}
		}

		if mind == -1 && wind == -1 && r.Body != nil {

			data := make([]byte, r.ContentLength)
			total, err := r.Body.Read(data)

			if err != nil {
				if err != io.EOF {
					log.Println("Request Read Body Error", err)
					return
				}
			}

			g.Set("Data", true)
			g.Set("Form", false)
			g.Set("Value", total)
			g.Set("Body", data)

		}
	}
}

//Dial beings the service connection
func (m *HTTPService) Dial() error {
	if m.cert != nil {
		// return http.ListenAndServe(m.GetPath(), http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		// 	m.Route.IssueRequestPath(r.URL.Path, func(pack *grids.GridPacket) {
		// 		pack.Set("Req", r)
		// 		pack.Set("Res", rw)
		// 		CollectHTTPBody(r, pack)
		// 	})
		// }))
		return http.ListenAndServe(m.GetPath(), http.HandlerFunc(m.ProcessPackets))
	}
	return http.ListenAndServeTLS(m.GetPath(), m.cert.Cert, m.cert.Key, http.HandlerFunc(m.ProcessPackets))
}

//ProcessPackets takes the req and response objects from the http server and wraps them in a grid packet
//for use in the service framework
func (m *HTTPService) ProcessPackets(rw http.ResponseWriter, r *http.Request) {
	m.Route.IssueRequestPath(r.URL.Path, func(pack *grids.GridPacket) {
		pack.Set("Req", r)
		pack.Set("Res", rw)
		CollectHTTPBody(r, pack)
	})
}

//WhenServiceJSON checks a gridpacket for the json flag and gets the body sorted from the
//http request for use
var WhenServiceJSON = func(g *grids.GridPacket, next func(li *arch.LinkDescriptor, g *grids.GridPacket)) {
	jsn, _ := g.Get("JSON").(bool)

	if !jsn {
		return
	}

	body, ok := g.Get("Body").([]byte)

	if !ok {
		return
	}

	li := new(arch.LinkDescriptor)
	err := json.Unmarshal(body, li)

	if err != nil {
		log.Fatal("Error occured when unmarshalling json to linkdescriptor", err, li)
		return
	}

	next(li, g)
}

//ExtractReqRes collects the request and response from a gridpacket
var ExtractReqRes = func(g *grids.GridPacket, next func(res http.ResponseWriter, req *http.Request)) {
	req, ok := g.Get("Req").(*http.Request)

	if !ok {
		return
	}

	res, ok := g.Get("Res").(http.ResponseWriter)

	if !ok {
		return
	}

	next(res, req)
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
			WhenServiceJSON(g, func(li *arch.LinkDescriptor, _ *grids.GridPacket) {
				sm.Register(li.Service, li)
				ExtractReqRes(g, func(res http.ResponseWriter, req *http.Request) {
					if sm.HasRegistered(li.Service) {
						res.WriteHeader(200)
					} else {
						res.WriteHeader(404)
					}
				})
			})
		}))
	}

	disc, err := sm.Select("discover")

	if err == nil {
		disc.Terminal().Any(grids.ByPackets(func(g *grids.GridPacket) {
			path, _ := g.Get("Pathways").([]string)
			service := path[0]

			ExtractReqRes(g, func(res http.ResponseWriter, req *http.Request) {
				li, err := sm.GetServiceProvider(service)

				if err != nil {
					log.Fatal("Unable to find service", service)
					res.WriteHeader(404)
					return
				}

				bin, err := json.Marshal(li)

				if err != nil {
					log.Fatal("Unable to jsonify service linkdescriptor: ", err, li)
					res.WriteHeader(404)
					return
				}

				res.Header().Set("Content-Type", "charset=utf-8;application/json")
				res.WriteHeader(200)
				res.Write(bin)

			})

		}))
	}

	unreg, err := sm.Select("unregister")

	if err == nil {
		unreg.Terminal().Only(grids.ByPackets(func(g *grids.GridPacket) {
			WhenServiceJSON(g, func(li *arch.LinkDescriptor, _ *grids.GridPacket) {
				sm.Unregister(li.Service, li)
				ExtractReqRes(g, func(res http.ResponseWriter, req *http.Request) {
					if !sm.HasRegistered(li.Service) {
						res.WriteHeader(200)
					} else {
						res.WriteHeader(404)
					}
				})
			})
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
