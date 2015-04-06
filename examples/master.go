package main

import (
	"fmt"

	"github.com/influx6/composelab"
)

func main() {

	master := composelab.NewMaster("127.0.0.1", 3002)
	master.Dial()

	fmt.Println("master-Path:", master.Location())

}
