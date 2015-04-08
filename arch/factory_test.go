package arch

import (
	"testing"

	"github.com/franela/goblin"
)

func TestFactory(t *testing.T) {
	g := goblin.Goblin(t)

	g.Describe("Create Factory", func() {

		fl := NewFactory()
		fl.Provide("flux", func(m *LinkDescriptor) (Linkage, error) {
			return Linkage(NewServiceLink(m)), nil
		})

		g.It("can i create a factory", func() {
			g.Assert(fl == nil).IsFalse()
		})

		g.It("do we have a flux generator", func() {
			g.Assert(fl.Has("flux")).IsTrue("factory has flux service link")
		})

		g.It("can i create a flux generator", func() {
			flx, err := fl.Resolve(NewDescriptor("flux", "drop", "0.0.0.0", 300, "0", ""))

			g.Assert(err == nil).IsTrue("error is nil")

			_, ok := flx.(*ServiceLink)
			g.Assert(ok).IsTrue("its a service link")
		})
	})
}
