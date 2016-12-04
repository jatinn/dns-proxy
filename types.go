package main

import (
	"fmt"
	"log"
	"net/http/httputil"
	"net/url"
	"strings"
)

type Config struct {
	Bind     string       `toml:"bind"`
	HTTPPort int          `toml:"http-port"`
	DNSPort  int          `toml:"dns-port"`
	TLD      string       `toml:"tld"`
	Routes   RoutingTable `toml:"routes"`
}

func NewConfig() *Config {
	return &Config{Routes: make(RoutingTable)}
}

// RoutingTable implements the cli.Generic interface so that we can build a
// map[string]string struct from command line arguments.
// String input of `src=dest,src1=dest1` gets parsed to build `{src: dest, src1: dest1}`
type RoutingTable map[string]string

func (r *RoutingTable) Set(value string) error {
	parts := strings.Split(value, ",")

	for _, pair := range parts {
		route := strings.Split(pair, "=")
		if len(route) == 1 {
			return fmt.Errorf("invalid format")
		}
		(*r)[route[0]] = route[1]
	}

	return nil
}

func (r *RoutingTable) String() string {
	if len(*r) == 0 {
		return ""
	}

	var items = make([]string, 0, len(*r))
	for k, v := range *r {
		items = append(items, fmt.Sprintf("%s = %s", k, v))
	}
	return fmt.Sprintf("Routes{%s}", strings.Join(items, ", "))
}

type ProxyRule struct {
	Name    string
	Address string
	proxy   *httputil.ReverseProxy
}

type ProxyRouter []*ProxyRule

func NewRouter(c *Config) *ProxyRouter {
	var router = make(ProxyRouter, 0, len(c.Routes))
	for k, v := range c.Routes {
		var address string
		route := fmt.Sprintf("%s.%s", k, c.TLD)
		if k == "" {
			route = c.TLD
		}

		if strings.HasPrefix(v, ":") {
			address = fmt.Sprintf("http://%s%s", c.Bind, v)
		} else if strings.HasPrefix(v, "http") {
			address = v
		} else {
			address = fmt.Sprintf("http://%s", v)
		}

		url, err := url.Parse(address)
		if err != nil {
			log.Fatalf("could not build target url for %s = %s: %v/n", route, address, err)
		}
		router = append(router, &ProxyRule{
			Name:    route,
			Address: address,
			proxy:   httputil.NewSingleHostReverseProxy(url),
		})
	}
	return &router
}

func (r *ProxyRouter) Match(host string) *httputil.ReverseProxy {
	for _, rule := range *r {
		if host == rule.Name || strings.HasSuffix(host, "."+rule.Name) {
			return rule.proxy
		}
	}
	return nil
}

func (r *ProxyRouter) String() string {
	var rules = make([]string, len(*r))
	for i, rule := range *r {
		rules[i] = fmt.Sprintf("*.%s => %s", rule.Name, rule.Address)
	}
	return strings.Join(rules, "\n")
}
