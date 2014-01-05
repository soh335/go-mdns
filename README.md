# go-mdns

## feature

* discover service by multicast udp
* only discover function

## usage

```go
client := new(mdns.Client)
client.Discover("_airplay._tcp.local.", func(msg *dns.Msg) {
        for _, rr := range msg.Extra {
                switch rr := rr.(type) {
                case *dns.A:
                        fmt.Println(rr.Header().Name, "=>", rr.A)
                default:
                }
        }
})
```

## see also

* https://github.com/soh335/go-dnssd ( dnssd implementation )
* https://github.com/miyagawa/AnyEvent-mDNS ( multicast udp implementation by perl )
* https://github.com/davecheney/mdns ( publish support )

