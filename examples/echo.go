package main

import (
	"fmt"

	"github.com/influx6/composelab/arch"
)

func main() {

	echo := arch.NewHttpSlave("echo", "127.0.0.1", 3001, "127.0.0.1", 3002)

	echo.Dial()
	echo.Discover("prime")
	echo.Discover("views")

	fmt.Println("echo-Path:", echo.Location())

}
