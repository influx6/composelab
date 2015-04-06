package main

import (
	"fmt"

	"github.com/influx6/composelab/links"
	"github.com/influx6/composelab/services"
)

func main() {

	master := links.NewHttpLinkage("127.0.0.1", 3002)
	prime := services.NewHttpSlave("prime", "127.0.0.1", 3003, master)

	prime.Dial()

	fmt.Println("prime-Path:", prime.Location())

}
