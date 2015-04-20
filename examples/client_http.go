package main

import (
	"log"
	"sync"

	"github.com/influx6/composelab/arch"
	"github.com/influx6/composelab/links"
)

func main() {

	wc := new(sync.WaitGroup)
	link := links.NewHTTPLink("flux", "127.0.0.1", 6300)

	link.Dial()

	wc.Add(3)

	dc := arch.NewDescriptor("udp", "fluv", "127.0.0.1", 3040, "Lagos", "upd4")

	link.Register("fluv", dc, func(d ...interface{}) {
		log.Println("registered:", d, dc)
		wc.Done()
	})

	link.Discover("fluv", func(target string, data interface{}, ud interface{}) {
		log.Println("target:", target, ud, dc)
		wc.Done()
	})

	link.Unregister("fluv", dc, func(d ...interface{}) {
		log.Println("unregistered:", d, dc)
		wc.Done()
	})

	wc.Wait()

}
