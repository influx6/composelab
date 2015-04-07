package main

import (
	"log"
	"sync"

	"github.com/influx6/composelab/links"
)

func main() {

	wc := new(sync.WaitGroup)
	link, err := links.NewUDPLink("go", "127.0.0.1", 3000)

	if err != nil {
		log.Println("Error occured:", err)
		return
	}

	link.Dial()

	wc.Add(1)

	link.Discover("/fluv", func(target string, data interface{}, ud interface{}) {
		log.Println("target:", target, data, ud)
		wc.Done()
	})

	wc.Wait()

}
