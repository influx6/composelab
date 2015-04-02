package main

import (
	"fmt"
	"github.com/influx6/composelab/arch"
)

func main() {

	prime := arch.NewSlave("prime", "127.0.0.1", 3001, "127.0.0.1", 3002)

	prime.Dial()

	fmt.Println("prime-Path:", prime.Location())

}
