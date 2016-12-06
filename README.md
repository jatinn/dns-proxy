# DNS-Proxy

A command line tool that aims to simplify configuring your development environment
for using multiple named domains. It starts a DNS and a reverse proxy server to
map routes such as app.localhost to their respective service port. To do this 
generally required running two services such as dnsmasq & nginx, however that
seemed like overkill for doing local development.

## Getting Started

### Installing

Make sure you have a working Go environment. Then just run the following:

```go get github.com/jatinn/dns-proxy```


### Configuring your development TLD

**Note** Steps are for Mac OSX.

While you are free to use whatever domain you choose, I would recommend using
one of the following restricted TLDs: [Go here to learn more](https://iyware.com/dont-use-dev-for-development/)

- test
- example
- invalid
- localhost

I will procced using `localhost` but the steps should be the same no matter
what tld you choose.

First we need to add an entry to map the localhost TLD to '127.0.0.1'. This
will route all traffic to the TLD whereas /etc/hosts does not support
wildcards so you would need to add an entry for each route you use.

```
sudo mkdir -v /etc/resolver
echo "nameserver 127.0.0.1" | sudo tee -a /etc/resolver/localhost
```

Next run dns-proxy mapping domains to their service ports. As an example I
have two services running on ports 8000 & 9000.

```
sudo ./dns-proxy --routes app1=:8000,app2=:9000 --tld localhost
```

Now I can make requests to app1.localhost and app2.localhost which get routed
to the correct services. `localhost` is the default option for the tld so that
flag could have been omitted in the command above.

![preview of dns-proxy routing traffic](https://cloud.githubusercontent.com/assets/5474746/20868990/05f4897a-ba37-11e6-88eb-dd644d194720.png)

#### Config File

All the options for the tool can be saved in a config file which should help if
you have many different routes you wish to use. Once you have your config file
saved you can run dns-proxy with the `--config` flag or set the `DNS_PROXY_CONFIG`
environment variable. 

The config file uses the [TOML](https://github.com/toml-lang/toml) format. See [sample_conf.toml](sample_conf.toml) for all the options.

## Usage Reference

```
NAME:
   dns-proxy - Simple DNS and proxy server for developing with named domains.

  DNS-Proxy aims to provide a simple way to route different domains/subdomains
  to services running locally. For example: app.localhost -> localhost:8080,
  redis.localhost -> 192.168.0.1:6379.

USAGE:
   dns-proxy [global options] command [command options] [arguments...]

VERSION:
   0.1.0

COMMANDS:
     help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --bind value       ip bound for internal routing (default: "127.0.0.1")
   --http-port value  port on which the proxy server will listen (default: 80)
   --dns-port value   port on which the DNS server will listen (default: 53)
   --tld value        tld to be used for development (default: "localhost")
   --routes value     comma-separated list of route pairs SOURCE=DEST.
                      The specified tld is appended to the SOURCE and the DEST
                      is host:port ex:
                      dns-proxy --routes 'app=:5000,api=192.168.0.1:3000'
                      which would proxy requests from *.app.localhost to
                      127.0.0.1:5000 and *.api.localhost to 192.168.0.1:3000
   --config FILE      load configuration from FILE [$DNS_PROXY_CONFIG]
   --help, -h         show help
   --version, -v      print the version
```

## TODO

- improve documentation
- testing
- tool name ???
- automate binary releases
- look into how to make it autostart on boot
