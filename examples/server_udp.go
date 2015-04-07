package main

import (
	"encoding/json"
	"log"
	"net"

	"github.com/influx6/composelab/arch"
)

func main() {

	udpAddr, err := net.ResolveUDPAddr("udp4", "127.0.0.1:3000")
	end := make(chan interface{})

	if err != nil {
		return
	}

	conn, err := net.ListenUDP("udp", udpAddr)

	if err != nil {
		return
	}

	buf := make([]byte, 1024)

	for {
		select {
		case <-end:
			log.Println("closing connection")
			conn.Close()
		default:
			len, addr, err := conn.ReadFromUDP(buf)

			if err != nil {
				log.Println("Error", err)
				end <- 1
				return
			}

			data := buf[:len]

			var jx = new(arch.UDPPack)
			err = json.Unmarshal(data, jx)

			log.Println("server received data:", addr, jx, err, string(data))

			conn.WriteTo(data, addr)

		}
	}

}
