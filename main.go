package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/miekg/dns"
)

func main() {
	// Using cli app as a parser to process the flags and build a config struct
	// which is passed to the supplied function.
	NewParser(runner).Run(os.Args)
}

func runner(c *Config) {
	log.Println("configuring app")
	var (
		dnsAddress  = net.JoinHostPort(c.Bind, strconv.Itoa(c.DNSPort))
		httpAddress = net.JoinHostPort(c.Bind, strconv.Itoa(c.HTTPPort))
		proxyIP     = net.ParseIP(c.Bind)
		routes      = NewRouter(c)
	)

	log.Printf("starting DNS to handle *.%s domains at %s\n", c.TLD, dnsAddress)
	dns.HandleFunc(c.TLD, func(w dns.ResponseWriter, r *dns.Msg) {
		m := new(dns.Msg)
		m.SetReply(r)
		m.Answer = append(m.Answer, &dns.A{
			Hdr: dns.RR_Header{Name: r.Question[0].Name, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 10},
			A:   proxyIP,
		})
		w.WriteMsg(m)
	})

	go serveDNS(dnsAddress, "tcp")
	go serveDNS(dnsAddress, "udp")

	log.Printf("starting HTTP proxy at %s\n", httpAddress)
	log.Printf("routing table: \n%s", routes)

	s := &http.Server{
		Addr:         httpAddress,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		Handler: func() http.HandlerFunc {
			return func(w http.ResponseWriter, req *http.Request) {

				// if request host is domain or subdomain match for any of the
				// supplied routes proxy the request over
				if proxy := routes.Match(req.Host); proxy != nil {
					proxy.ServeHTTP(w, req)
					return
				}

				w.WriteHeader(http.StatusNotFound)
				fmt.Fprintf(w, "no route setup to handle %s\ncurrent routes: \n%v", req.Host, routes)
			}
		}(),
	}

	if err := s.ListenAndServe(); err != nil {
		fmt.Printf("failed to setup the proxy server: %v\n", err)
	}
}

func serveDNS(address, net string) {
	server := &dns.Server{Addr: address, Net: net, TsigSecret: nil}
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("failed to setup the %s DNS server: %v\n", net, err)
	}
}
