package main

import (
	"flag"
	"fmt"
	"github.com/miekg/dns"
	"log"
	"regexp"
	"strconv"
)

var interceptDomain = "ip.net."
var ip, _ = regexp.Compile("(.*\\.)?([0-2]?\\d?\\d\\.[0-2]?\\d?\\d\\.[0-2]?\\d?\\d\\.[0-2]?\\d?\\d)\\.ip\\.net")
var ttl = "3600"

var remoteDns string

func interceptDnsRequest(w dns.ResponseWriter, m *dns.Msg) {
	r := new(dns.Msg)
	r.SetReply(m)
	r.Compress = false

	switch m.Opcode {
	case dns.OpcodeQuery:
		for _, q := range m.Question {
			switch q.Qtype {
			case dns.TypeA:
				log.Printf("Intercept A %s\n", q.Name)
				ip := ip.FindStringSubmatch(q.Name)[2]
				a, e := dns.NewRR(fmt.Sprintf("%s %s A %s", q.Name, ttl, ip))
				if e != nil {
					log.Fatalf("Failed to handle request: %s\n ", e.Error())
				}
				log.Printf("Answer %s\n", a)
				r.Answer = append(r.Answer, a)
				break
			case dns.TypeAAAA:
				log.Printf("Intercept AAAA %s. This is currently not supported.\n", q.Name)
				r.Rcode = dns.RcodeNameError
				break
			default:
				log.Printf("Qtype %d net set up.", q.Qtype)
				r.Rcode = dns.RcodeServerFailure
			}
		}
		break
	default:
		log.Printf("Opcode %d not set up.", m.Opcode)
	}

	err := w.WriteMsg(r)
	if err != nil {
		log.Fatalf("Failed to handle request: %s\n ", err.Error())
	}
}

func handleDnsRequest(w dns.ResponseWriter, m *dns.Msg) {
	r := new(dns.Msg)
	r.SetReply(m)
	r.Compress = false

	n := new(dns.Msg)
	n.Id = dns.Id()
	n.RecursionDesired = true
	n.Question = r.Question
	c := new(dns.Client)
	a, t, e := c.Exchange(n, remoteDns)
	if e != nil {
		log.Fatalf("Failed to forward request: %s\n", e.Error())
	}
	log.Printf("Forward response: %s (%d)\n", a.Answer, t)
	r.Answer = a.Answer

	err := w.WriteMsg(r)
	if err != nil {
		log.Fatalf("Failed to handle request: %s\n ", err.Error())
	}
}

func main() {
	var remotePort int
	remoteIp := flag.String("remote-dns-ip", "1.1.1.1", "ip address of the next dns resolver")
	flag.IntVar(&remotePort, "remote-dns-port", 53, "port of the next dns resolver")
	flag.Parse()

	remoteDns = *remoteIp + ":" + strconv.Itoa(remotePort)
	log.Printf("Forwarding non intercepted requests to %s\n", remoteDns)

	dns.HandleFunc(interceptDomain, interceptDnsRequest)
	dns.HandleFunc(".", handleDnsRequest)

	server := dns.Server{Addr: "0.0.0.0:53", Net: "udp"}
	err := server.ListenAndServe()
	defer server.Shutdown()
	if err != nil {
		log.Fatalf("Failed to start server (check if you have permission to listen on specific port): %s\n ", err.Error())
	}
}
