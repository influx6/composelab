package links

import (
	"fmt"
	"log"
	"net"
	"testing"

	"github.com/franela/goblin"
)

func CreateUDPServer(addr string, end chan struct{}) {
	udpAddr, err := net.ResolveUDPAddr("udp4", addr)

	if err != nil {
		return
	}

	conn, err := net.ListenUDP("udp", udpAddr)
	buf := make([]byte, 1024)

	for {
		select {
		case <-end:
			log.Println("closing connection")
			conn.Close()
		default:
			len, addr, err := conn.ReadFromUDP(buf)
			data := buf[:len]
			log.Println("server received data:", data, addr)

			if err != nil {
				log.Println("Error", err)
			}

		}
	}

}

func TestUDPClient(t *testing.T) {
	g := goblin.Goblin(t)
	addr := "127.0.0.1"
	port := 3000

	closer := make(chan struct{})
	go CreateUDPServer(fmt.Sprintf("%s:%d", addr, port), closer)

	link, err := NewUDPLink("go", addr, port)

	g.Describe("Udp link", func() {

		g.It("can i create a udp link", func() {
			g.Assert(err).Equal(nil)
			g.Assert(link != nil).IsTrue("link is not false")
		})

		g.It("can i send data over the link", func() {
			link.Conn.Write([]byte("solo"))
		})

	})
}
