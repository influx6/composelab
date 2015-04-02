package main

import (
	"fmt"
	"github.com/influx6/composelab/arch"
)

func main() {

	master := arch.NewMaster("127.0.0.1", 3002)
	master.Dial()

	fmt.Println("master-Path:", master.Location())

}
