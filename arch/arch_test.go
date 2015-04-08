package arch

import (
	"testing"

	"github.com/franela/goblin"
)

func TestArch(t *testing.T) {
	g := goblin.Goblin(t)

	g.Describe("Base Architecture", func() {

		sd := NewDescriptor("uup", "load", "127.0.0.1", 3000, "0", "")
		sil := NewServiceLink(sd)

		g.It("can i create a linkage", func() {
			g.Assert(sil == nil).IsFalse("new link is not false")
			g.Assert(sil.GetPort()).Eql(3000)
			g.Assert(sil.GetAddress()).Eql("127.0.0.1")
			g.Assert(sil.GetPath()).Eql("127.0.0.1:3000")
		})

		g.It("can i create a service", func() {
			d := NewDescriptor("uup", "flux", "0.0.0.0", 3001, "0", "")
			sm := NewService(d, Linkage(sil))
			g.Assert(sm.Master).Eql(sil)
			g.Assert(d.Address).Eql(sm.GetAddress())
			g.Assert(d.Port).Eql(sm.GetPort())
		})
	})
}
