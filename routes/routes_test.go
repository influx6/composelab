package routes

import (
	"log"
	"testing"

	. "github.com/franela/goblin"
	"github.com/influx6/goutils"
	"github.com/influx6/grids"
)

func TestRouteKeeper(t *testing.T) {
	gob := Goblin(t)

	gob.Describe("RouteKeeper specification", func() {

		gob.It("can i create a routekeeper", func() {
			k := NewRouteKeeper(0, nil)
			gob.Assert(k != nil).IsTrue()
		})

		gob.It("can i fail a routekeeper", func(done Done) {
			k := NewRouteKeeper(1, func(f *grids.GridPacket) {
				gob.Assert(f != nil).IsTrue("we got a gridPacket")
				log.Println("i just failed in 1ms")
				done()
			})

			k.Secure(grids.NewPacket())
		})
	})

}

func TestRoute(t *testing.T) {
	gob := Goblin(t)
	app := NewRoutes("app", true) //=> creates a /app route
	app.Branch(`{id:[\d+]}`, true)
	app.Branch("logs/realtime", true) //=> creates a /app/log/realtime route

	gob.Describe("testing route making type", func() {
		gob.It("can i create a app route", func() {
			rtype := RouteTypes(app)
			gob.Assert(rtype).Equal(app)
		})
	})

	gob.Describe("can i create under-routes", func() {
		gob.It("can i create /:id route", func() {
			gob.Assert(app.Has("id")).IsTrue()
		})

		gob.It("can i create /logs/realtime route", func() {
			gob.Assert(app.Has("logs/realtime")).IsTrue()
		})

		gob.It("can i select the realtime route from the root", func() {
			_, err := app.Select("logs/realtime")
			gob.Assert(err == nil).IsTrue()
		})

	})

	gob.Describe("can i send a request for", func() {
		gob.It("sending request for /logs", func(done Done) {
			logs, err := app.Select("logs")

			gob.Assert(err == nil).IsTrue()

			logs.Terminal().Only(func(req interface{}, _ *RouteFinalizer) {
				_, ok := req.(*grids.GridPacket)
				gob.Assert(ok).IsTrue()
				done()
			})

			pack := grids.NewPacket()
			pack.Set("Pathways", []string{"app", "logs"})
			app.InSend("Request", pack)
		})

		gob.It("sending request for /id", func(d Done) {
			id, err := app.Select("id")

			gob.Assert(err == nil).IsTrue()

			id.Terminal().Any(func(req interface{}, r *RouteFinalizer) {
				f, ok := req.(*grids.GridPacket)
				gob.Assert(ok).IsTrue()
				params, ok := f.Get("Params").(*goutils.Map)
				gob.Assert(params.Get("id")).Equal("4")

				r.Effect(func(r *grids.GridPacket) {
					gob.Assert(f).Equal(r)
					d()
				})

			})

			app.IssueRequestPath("app/4", nil)
		})

	})

}
