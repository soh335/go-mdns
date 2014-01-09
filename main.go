package mdns

import (
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

	defer conn.Close()
	c.conn = conn

	out, err := m.Pack()

	if err != nil {
		panic(err)
	}

	_, err = conn.WriteToUDP(out, addr)
	if err != nil {
		panic(err)
	}

	c.handleReceiveMsg(domain, cb)
}

func (c *Client) handleReceiveMsg(domain string, cb func(*dns.Msg)) {
	timeout := defaultTimeout
	if c.Timeout != 0 {
		timeout = c.Timeout
	}
	timer := time.After(timeout)
	msgChan := make(chan *dns.Msg)
	closed := false
	doneChan := make(chan bool)

	go func() {
		for {
			if _, msg, err := c.readUDP(); err != nil {
				if closed {
					doneChan <- true
					return
				} else {
					panic(err)
				}
			} else {
				msgChan <- msg
			}
		}
	}()

	found := make(map[string]*dns.Msg)

	for {
		select {
		case <-timer:
			closed = true
			c.conn.Close()
			break
		case msg := <-msgChan:
			for _, rr := range msg.Answer {
				switch rr := rr.(type) {
				case *dns.PTR:
					if rr.Header().Name != domain {
						continue
					}
					ptr := rr.Ptr
					if _, ok := found[ptr]; ok {
						break
					}
					found[ptr] = msg
					cb(msg)
				}
			}
		case <-doneChan:
			return
		}
	}

}

func (c *Client) readUDP() (*net.UDPAddr, *dns.Msg, error) {
	in := make([]byte, dns.DefaultMsgSize)
	read, addr, err := c.conn.ReadFromUDP(in)
	if err != nil {
		return nil, nil, err
	}

	var readMsg dns.Msg
	if err := readMsg.Unpack(in[:read]); err != nil {
		return nil, nil, err
	}

	return addr, &readMsg, nil
}
