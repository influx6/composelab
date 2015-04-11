package main

import (
	"log"

	"github.com/influx6/composelab/services"
)

func main() {

	// wc := new(sync.WaitGroup)
	ud, err := services.NewUDPService("flux", "127.0.0.1", 6300, nil)

	if err != nil {
		log.Fatal("Error occured in creating service", err, ud)
	}

	// wc.Add(1)
	ud.Dial()

	// wc.Wait()
}
