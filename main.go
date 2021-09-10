package main

import (
	"flag"
	"fmt"

	"github.com/arturoeanton/go-teleport/network"
)

var mapMirrors = make(map[string]*network.Mirror)

var name = flag.String("name", "mirrot1", "Name of the mirror")
var protocol = flag.String("protocol", "tcp", "Protocol to use")
var addr1 = flag.String("addr1", "8081", "Address to use")
var addr2 = flag.String("addr2", "8082", "Address to use")

func main() {
	fmt.Println("Pivot v0.0.1")
	flag.Parse()

	if _, ok := mapMirrors[*name]; !ok {
		mirror := &network.Mirror{
			Addr1:    *addr1,
			Addr2:    *addr2,
			Protocol: *protocol,
		}
		mapMirrors[*name] = mirror
		go mirror.Start()
	}
	<-make(chan int)

}
