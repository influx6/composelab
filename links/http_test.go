package links

import (
	"log"
	"testing"

	"github.com/franela/goblin"
)

func TestHTTPClient(t *testing.T) {
	g := goblin.Goblin(t)

	g.Describe("Create HttpClient", func() {
		log.Println("new client")
	})
}
