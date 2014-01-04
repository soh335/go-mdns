package main

import (
	"fmt"
	"github.com/miekg/dns"
	"github.com/soh335/go-mdns"
	"os"
)

func main() {
	client := new(mdns.Client)
	client.Discover(os.Args[1], func(msg *dns.Msg) {
		for _, rr := range msg.Extra {
			switch rr := rr.(type) {
			case *dns.A:
				fmt.Println(rr.Header().Name, "=>", rr.A)
			default:
			}
		}
	})
}
