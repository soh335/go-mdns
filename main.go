package mdns

import (
	"fmt"
	"github.com/miekg/dns"
	"net"
	"time"
)

const defaultTimeout time.Duration = time.Second * 3

type Client struct {
	Timeout time.Duration
	conn    *net.UDPConn
}

func (c *Client) Discover(domain string, cb func(*dns.Msg)) {
	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(domain), dns.TypePTR)
	m.RecursionDesired = true

	addr := &net.UDPAddr{
		IP:   net.ParseIP("224.0.0.251"),
		Port: 5353,
	}

	conn, err := net.ListenMulticastUDP("udp4", nil, addr)

	if err != nil {
		panic(err)
	}

	c.conn = conn

	out, err := m.Pack()

	if err != nil {
		panic(err)
	}

	_, err = conn.WriteToUDP(out, addr)
	if err != nil {
		panic(err)
	}

	c.handleReceiveMsg(cb)
}

func (c *Client) handleReceiveMsg(cb func(*dns.Msg)) {
	timeout := defaultTimeout
	if c.Timeout != 0 {
		timeout = c.Timeout
	}
	timer := time.After(timeout)
	msgChan := make(chan *dns.Msg)

	go func() {
		for {
			_, msg := c.read()
			msgChan <- msg
		}
	}()

	found := make(map[string]*dns.Msg)

	for {
		select {
		case <-timer:
			return
		case msg := <-msgChan:
			if len(msg.Answer) < 1 {
				continue
			}
			for _, rr := range msg.Answer {
				switch rr := rr.(type) {
				case *dns.PTR:
					ptr := rr.Ptr
					if _, ok := found[ptr]; ok {
						continue
					}
					found[ptr] = msg
					cb(msg)
				}
			}

		}
	}
}

func (c *Client) read() (*net.UDPAddr, *dns.Msg) {
	in := make([]byte, 1024)
	read, addr, err := c.conn.ReadFromUDP(in)
	if err != nil {
		panic(err)
	}

	var readMsg dns.Msg
	if err := readMsg.Unpack(in[:read]); err != nil {
		fmt.Println(&readMsg)
		panic(err)
	}

	return addr, &readMsg
}
