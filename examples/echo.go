package main

import (
	"github.com/influx6/composelab/links"
	"github.com/influx6/composelab/services"
)

func main() {

	master := links.NewHttpLinkage("127.0.0.1", 3002)
	echo := services.NewHttpSlave("echo", "127.0.0.1", 3001, master)

	echo.Dial()
	echo.Discover("prime")
	echo.Discover("views")

}
