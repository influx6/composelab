//Packege Composelab/Routes
//combines grids and a clever api to create branchable routes and paths

package routes

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/influx6/evroll"
	"github.com/influx6/goutils"
	"github.com/influx6/grids"
	"github.com/influx6/immute"
	"github.com/influx6/reggy"
)

//RouteFinalizer takes the packets recieved from the routes and packs them into
//a stack for processing,these stack can be cleared when done or allowed to remain
//during the lifetime of the route
type RouteFinalizer struct {
	flush bool
	Box   *immute.Sequence
	Added *evroll.EventRoll
}

//RouteKeeper takes a gridpacket request then handles it into a blocking chan which
//lives for a specific period of time-limit,if its time limit is given or else uses
//a default time limit at which it discards and dies off ,closing its chan
type RouteKeeper struct {
	block         chan *grids.GridPacket
	defaultAction func(*grids.GridPacket)
	timeout       time.Duration
	buffered      bool
	done          bool
}

//Release returns grids.GridPacket,Error
//i.e if the timeout is not over and the keeper has not expired
//then it returns a (gridPacket,nil) else if it has been released then
//it returns a (nil,ErrorDone) else a (nil,Error) indicating
//timeout and default action processed
func (rk *RouteKeeper) Release(cbx func(*grids.GridPacket)) error {
	p := <-rk.block

	if p == nil {
		return errors.New("RouteKeeper Expired")
	}

	if rk.buffered {
		rk.block <- p
	} else {
		rk.done = true
	}

	cbx(p)

	return nil
}

// //Close calls a close on the keepers channel
// func (rk *RouteKeeper) Close() {
// 	if !rk.done {
// 		defer close(rk.block)
// 	}
// }

//Secure collects the gridPacket to be kept by the routekeeper
func (rk *RouteKeeper) Secure(g *grids.GridPacket) {
	go func() {
		go func() {
			rk.block <- g
		}()

		<-time.After(rk.timeout)

		if !rk.buffered {
			defer close(rk.block)
			if rk.defaultAction != nil {
				rk.defaultAction(<-rk.block)
			} else {
				<-rk.block
			}
			rk.done = true
		}
	}()
}

//NewRouteKeeper returns a new routekeeper that locks its gridpacket and waits for a timeout to
//kill the request
func NewRouteKeeper(timeout int, fail func(*grids.GridPacket)) *RouteKeeper {

	var mc chan *grids.GridPacket
	var ms time.Duration
	var buf bool

	if timeout <= 0 {
		mc = make(chan *grids.GridPacket, 1)
		buf = true
	} else {
		mc = make(chan *grids.GridPacket)
		buf = false
	}

	ms = time.Duration(timeout) * time.Millisecond

	return &RouteKeeper{
		mc,
		fail,
		ms,
		buf,
		false,
	}

}

//NewRouteFinalizer returns a new RouteFinalizer
func NewRouteFinalizer(f bool, fail func(g *grids.GridPacket)) *RouteFinalizer {
	rf := &RouteFinalizer{
		f,
		immute.CreateList(make([]interface{}, 0)),
		evroll.NewEvent("Added"),
	}

	rf.Added.Listen(grids.ByPackets(func(data *grids.GridPacket) {
		ms, ok := data.Get("timeout").(int)

		if !ok {
			ms = 6
		}

		kp := NewRouteKeeper(ms, fail)
		rf.Box.Add(kp, nil)
		kp.Secure(data)
	}))

	return rf
}

//Drop adds a new request gridPacket to the routefinalizer stack for treatment
func (rf *RouteFinalizer) Drop(g *grids.GridPacket) {
	rf.Added.Emit(g)
}

//Walk calls a function through the packets stored within the streampack
//for the routefinalizer and watches for new packets arrival for response
func (rf *RouteFinalizer) Walk(cb func(g *grids.GridPacket), fcb func()) {
	rf.Box.Each(func(data interface{}, key interface{}) interface{} {
		grids.ByPackets(func(r *grids.GridPacket) {
			cb(r)
		})(data)
		return nil
	}, func(_ int, _ interface{}) {
		if fcb != nil {
			fcb()
		}

		if rf.flush {
			rf.Box.Clear()
		}
	})
}

//Effect calls walk to perform the function assigned to handle all packets
func (rf *RouteFinalizer) Effect(cb func(g *grids.GridPacket)) {
	rf.Walk(cb, nil)
}

//Callable represents a func with a interface as argument
type Callable func(data interface{}, r *RouteFinalizer)

//RouteTypes is the standard defining interface for all routables
type RouteTypes interface {
	Select(string) (*Routes, error)
	Branch(string, bool)
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
	Loose  *RouteFinalizer //is emitted when path is within terminal's route
	Strict *RouteFinalizer //is emitted only when path is terminal's route
	// Route      *Routes
}

//Any adds a callback to listen to all events passing the AllEvents handler
func (t *Terminal) Any(c Callable) {
	t.Loose.Added.Listen(func(d interface{}) {
		c(d, t.Loose)
	})
}

//Only adds a callback to listen to events on the OnlyEvent handler
func (t *Terminal) Only(c Callable) {
	t.Strict.Added.Listen(func(d interface{}) {
		c(d, t.Strict)
	})
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
func (r *Routes) IssueRequestPath(path string, before func(*grids.GridPacket)) {
	pack := grids.NewPacket()
	pack.Set("Pathways", goutils.SplitPatternAndRemovePrefix(path))
	//send off the request
	if before != nil {
		before(pack)
	}
	r.InSend("Request", pack)
}

//IssueRequestPacket takes a string path and a packet and pushes off as a request
func (r *Routes) IssueRequestPacket(path string, pack *grids.GridPacket) {
	//add the Pathways meta data
	pack.Set("Pathways", goutils.SplitPatternAndRemovePrefix(path))
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
func (r *Routes) Branch(path string, drop bool) {
	if path == "" {
		return
	}

	parts := goutils.SplitPattern(path)

	if len(parts) <= 0 {
		return
	}

	first := parts[0]
	cur := NewRoutes(first, drop)
	rem := parts[1:]

	r.Links.Set(cur.Path, cur)
	r.OutBind("All", cur.In("Request"))

	if len(rem) <= 0 {
		return
	}

	cur.Branch(strings.Join(rem, "/"), drop)
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
func NewTerm(drop bool) *Terminal {
	term := &Terminal{
		NewRouteFinalizer(drop),
		NewRouteFinalizer(drop),
		// route,
	}
	return term
}

//NewRoutes returns a new route for a specific path
func NewRoutes(path string, drop bool) *Routes {
	matcher := reggy.GenerateClassicMatcher(path)
	return NewRoutesBy(matcher.Original, matcher, drop)
}

//NewRoutesBy returns a new route for a specific path but allows giving your own matcher
func NewRoutesBy(path string, pattern *reggy.ClassicMatcher, drop bool) *Routes {
	r := &Routes{
		grids.NewGrid("Composelab/Routes/" + path),
		path,
		pattern,
		goutils.NewMap(),
		NewTerm(drop),
	}
	// r.Term = NewTerm(r)

	r.NewIn("Request")

	//rejected request
	r.NewOut("Bad")

	//these handle internal request strictness(is it matching or only matches)
	r.NewOut("Only")
	r.NewOut("All")

	r.OrOut("All", func(p *grids.GridPacket) {
		r.Term.Loose.Drop(p)
	})

	r.OrOut("Only", func(p *grids.GridPacket) {
		r.Term.Strict.Drop(p)
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
