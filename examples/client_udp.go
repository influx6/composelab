package main

import (
	"log"
	"sync"

	"github.com/influx6/composelab/links"
)

func main() {

	wc := new(sync.WaitGroup)
	link, err := links.NewUDPLink("flux", "127.0.0.1", 6300)

	if err != nil {
		log.Println("Error occured:", err)
		return
	}

	link.Dial()

	wc.Add(1)

	link.Discover("fluv", func(target string, data interface{}, ud interface{}) {
		log.Println("target:", target, data, ud)
		wc.Done()
	})

	wc.Wait()

}
