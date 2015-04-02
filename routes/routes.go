//Packege Composelab/Routes
//combines grids and a clever api to create branchable routes and paths

package routes

import (
	"fmt"
	"strings"

	"github.com/influx6/evroll"
	"github.com/influx6/goutils"
	"github.com/influx6/grids"
	"github.com/influx6/reggy"
)

//Callable represents a func with a interface as argument
type Callable func(data interface{})

//RouteTypes is the standard defining interface for all routables
type RouteTypes interface {
	Select(string) (*Routes, error)
	Branch(string)
	Divert(*Routes) *Routes
	Terminal() *Terminal
}

//TerminalType describes the set of terminal must functions
type TerminalType interface {
	Only(Callable)
	Any(Callable)
}

//Terminal defines a means of providing events that simplifies the
//attaching for routes
type Terminal struct {
	OnlyEvents *evroll.EventRoll
	AllEvents  *evroll.EventRoll
	// Route      *Routes
}

//Any adds a callback to listen to all events passing the AllEvents handler
func (t *Terminal) Any(c evroll.Callabut) {
	t.AllEvents.Listen(c)
}

//Only adds a callback to listen to events on the OnlyEvent handler
func (t *Terminal) Only(c evroll.Callabut) {
	t.OnlyEvents.Listen(c)
}

//Routes is the base struct for defining interlinking routes
type Routes struct {
	*grids.Grid
	Path    string
	Pattern *reggy.ClassicMatcher
	Links   *goutils.Map
	Term    *Terminal
}

//IssueRequestPath takes a string path and a data to issue an auto packet request
//payload
func (r *Routes) IssueRequestPath(path string, head interface{}) *grids.GridPacket {
	pack := grids.NewPacket()
	pack.Set("Pathways", goutils.SplitPattern(path))
	pack.Set("Payload", head)
	//send off the request
	r.InSend("Request", pack)
	return pack
}

//IssueRequestPacket takes a string path and a packet and pushes off as a request
func (r *Routes) IssueRequestPacket(path string, pack *grids.GridPacket) {
	//add the Pathways meta data
	pack.Set("Pathways", goutils.SplitPattern(path))
	//send off the request
	r.InSend("Request", pack)
}

//IssueRequest takes a packet and pushes off as a request if it has
//the 'Pathways' meta path list else returns an err if it does not
func (r *Routes) IssueRequest(pack *grids.GridPacket) error {
	if !pack.Has("Pathways") {
		return fmt.Errorf("packet has no %s meta", "Pathways")
	}
	//send off the request
	r.InSend("Request", pack)
	return nil
}

//Divert takes another member route and when these routes rejects
//requests its sent to this,it returns the supplied *Routes for chaining
func (r *Routes) Divert(rw *Routes) *Routes {
	r.OutBind("Bad", rw.In("Request"))
	return rw
}

//Branch defines the member function of a route that defines the set
//of given pathways
func (r *Routes) Branch(path string) {
	if path == "" {
		return
	}

	parts := goutils.SplitPattern(path)

	if len(parts) <= 0 {
		return
	}

	first := parts[0]
	cur := NewRoutes(first)
	rem := parts[1:]

	r.Links.Set(cur.Path, cur)
	r.OutBind("All", cur.In("Request"))

	if len(rem) <= 0 {
		return
	}

	cur.Branch(strings.Join(rem, "/"))
}

//Has returns true or false if a route exists
func (r *Routes) Has(path string) bool {
	parts := goutils.SplitPattern(path)

	if len(parts) <= 0 {
		return false
	}

	first := parts[0]

	if !r.Links.Has(first) {
		return false
	}

	rem := parts[1:]
	if len(rem) <= 0 {
		return true
	}

	cur, ok := r.Links.Get(first).(*Routes)

	if !ok {
		return false
	}

	return cur.Has(strings.Join(rem, "/"))
}

//Select returns the route which has the specific pathway given
func (r *Routes) Select(path string) (*Routes, error) {
	parts := goutils.SplitPattern(path)

	if len(parts) <= 0 {
		return r, nil
	}

	first := parts[0]

	if !r.Links.Has(first) {
		return nil, fmt.Errorf("route not found %s", first)
	}

	cur, ok := r.Links.Get(first).(*Routes)

	if !ok {
		return nil, nil
	}

	rem := parts[1:]

	if len(rem) <= 0 {
		return cur, nil
	}

	return cur.Select(strings.Join(rem, "/"))
}

//Terminal returns the *Terminal set for these Route
func (r *Routes) Terminal() *Terminal {
	return r.Term
}

//NewTerm returns a terminal connected to a specific route
func NewTerm() *Terminal {
	term := &Terminal{
		evroll.NewEvent("Only"),
		evroll.NewEvent("Any"),
		// route,
	}
	return term
}

//NewRoutes returns a new route for a specific path
func NewRoutes(path string) *Routes {
	matcher := reggy.GenerateClassicMatcher(path)
	return NewRoutesBy(matcher.Original, matcher)
}

//NewRoutesBy returns a new route for a specific path but allows giving your own matcher
func NewRoutesBy(path string, pattern *reggy.ClassicMatcher) *Routes {
	r := &Routes{
		grids.NewGrid("Composelab/Routes/" + path),
		path,
		pattern,
		goutils.NewMap(),
		NewTerm(),
	}
	// r.Term = NewTerm(r)

	r.NewIn("Request")

	//rejected request
	r.NewOut("Bad")

	//these handle internal request strictness(is it matching or only matches)
	r.NewOut("Only")
	r.NewOut("All")

	r.OrOut("All", func(p *grids.GridPacket) {
		r.Term.AllEvents.Emit(p)
	})

	r.OrOut("Only", func(p *grids.GridPacket) {
		r.Term.OnlyEvents.Emit(p)
	})

	r.AndIn("Request", func(p *grids.GridPacket, next func(f *grids.GridPacket)) {
		if !p.Has("Pathways") {
			return
		}

		paths, ok := p.Get("Pathways").([]string)

		if len(paths) <= 0 {
			r.OutSend("Bad", p)
			return
		}

		if !ok {
			r.OutSend("Bad", p)
			return
		}

		first := paths[0]

		if !r.Pattern.Validate(first) {
			r.OutSend("Bad", p)
			next(p)
			return
		}

		if !p.Has("Params") {
			p.Set("Params", goutils.NewMap())
		}

		params, ok := p.Get("Params").(*goutils.Map)

		if ok {
			params.Set(r.Path, first)
		}

		rem := paths[1:]
		p.Set("Pathways", rem)

		if len(rem) <= 0 {
			r.OutSend("Only", p)
		}

		r.OutSend("All", p)

		next(p)
	})

	return r
}
