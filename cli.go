package main

import (
	"errors"
	"log"
	"os"

	"github.com/BurntSushi/toml"

	"gopkg.in/urfave/cli.v1"
)

func NewParser(fn func(*Config)) *cli.App {
	var cfg = NewConfig()

	app := cli.NewApp()
	app.Version = "0.1.0"
	app.Name = "dns-proxy"
	app.Usage = `Simple DNS and proxy server for developing with named domains.

	DNS-Proxy aims to provide a simple way to route different domains/subdomains
	to services running locally. For example: app.localhost -> localhost:8080,
	redis.localhost -> 192.168.0.1:6379.`

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "bind",
			Value:       "127.0.0.1",
			Usage:       "ip bound for internal routing",
			Destination: &cfg.Bind,
		},
		cli.IntFlag{
			Name:        "http-port",
			Value:       80,
			Usage:       "port on which the proxy server will listen",
			Destination: &cfg.HTTPPort,
		},
		cli.IntFlag{
			Name:        "dns-port",
			Value:       53,
			Usage:       "port on which the DNS server will listen",
			Destination: &cfg.DNSPort,
		},
		cli.StringFlag{
			Name:        "tld",
			Value:       "localhost",
			Usage:       "tld to be used for development",
			Destination: &cfg.TLD,
		},
		cli.GenericFlag{
			Name:  "routes",
			Value: &cfg.Routes,
			Usage: `comma-separated list of route pairs SOURCE=DEST.
	The specified tld is appended to the SOURCE and the DEST
	is host:port ex:
	dns-proxy --routes 'app=:5000,api=192.168.0.1:3000'
	which would proxy requests from *.app.localhost to
	127.0.0.1:5000 and *.api.localhost to 192.168.0.1:3000`,
		},
		cli.StringFlag{
			Name:   "config",
			Value:  "",
			Usage:  "load configuration from `FILE`",
			EnvVar: "DNS_PROXY_CONFIG",
		},
	}

	app.Action = func(c *cli.Context) error {
		if configfile := c.String("config"); configfile != "" {
			log.Printf("loading from config: %s\n", configfile)

			if _, err := os.Stat(configfile); err != nil {
				if os.IsNotExist(err) {
					log.Fatalln("config file does not exist")
				} else {
					log.Fatalf("error checking for config file: %v\n", err)
				}
			}

			if _, err := toml.DecodeFile(configfile, &cfg); err != nil {
				log.Fatalf("error decoding config file: %v\n", err)
			}
		}

		// At this point all arguments have been parsed and config contains the
		// supplied values or defaults. We now hand off the config to the main
		// function.
		if len(cfg.Routes) == 0 {
			return errors.New("no routes provided")
		}
		fn(cfg)
		return nil
	}

	return app
}
