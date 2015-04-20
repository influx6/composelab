package main

import "github.com/influx6/composelab/services"

func main() {

	// wc := new(sync.WaitGroup)
	htc := services.NewHTTPService("flux", "127.0.0.1", 6300, nil)

	// wc.Add(1)
	htc.Dial()

	// wc.Wait()
}
