package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	clc "github.com/CenturyLinkCloud/clc-sdk"
	"github.com/CenturyLinkCloud/clc-sdk/api"
	"github.com/CenturyLinkCloud/clc-sdk/server"
	"github.com/CenturyLinkCloud/clc-sdk/status"
)

const Version = "0.1"

func supportedProtocol(proto string) bool {
	proto = strings.ToLower(proto)
	switch proto {
	case
		"tcp",
		"udp",
		"icmp":
		return true
	}
	return false
}

func main() {
	un := flag.String("username", "", "clc username")
	pw := flag.String("password", "", "clc password")
	sid := flag.String("server", "", "server id")
	spec := flag.String("ports", "", "ports to open")
	intl := flag.String("internal", "", "(optional) internal ip")
	flag.Parse()
	if *un == "" {
		log.Panic("missing flag -username")
	}
	if *pw == "" {
		log.Panic("missing flag -password")
	}
	if *spec == "" {
		log.Panic("missing flag -ports")
	}
	// when not passed, use local hostname
	if *sid == "" {
		*sid, _ = os.Hostname()
	}
	if *intl != "" {
		log.Printf("Allocating on internal IP: %v", *intl)
	}

	config, _ := api.NewConfig(*un, *pw)
	config.UserAgent = fmt.Sprintf("public_ip: %s", Version)
	client := clc.New(config)
	if err := client.Authenticate(); err != nil {
		log.Panicf("Failed to auth: %v", err)
	}
	pubip := server.PublicIP{}
	ports := make([]server.Port, 0)
	for _, s := range strings.Split(*spec, " ") {
		x := strings.Split(s, "/")
		portrange, proto := x[0], x[1]
		if !supportedProtocol(proto) {
			log.Panicf("Unsupported protocol: %v", proto)
		}
		fromto := strings.Split(portrange, "-")
		var i, j int
		if len(fromto) > 1 {
			i, _ = strconv.Atoi(fromto[0])
			j, _ = strconv.Atoi(fromto[1])
		} else {
			i, _ = strconv.Atoi(fromto[0])
			j = -1
		}
		port := server.Port{
			Port:     i,
			Protocol: proto,
		}
		if j != -1 {
			port.PortTo = j
		}
		ports = append(ports, port)
	}
	pubip.Ports = ports

	var addr string
	var st *status.Status
	var svr *server.Response
	svr, err := client.Server.Get(*sid)
	if err != nil {
		log.Panicf("Failed fetching server: %v - %v", *sid, err)
	}

	for _, ip := range svr.Details.IPaddresses {
		addr = ip.Public
		pubip.InternalIP = ip.Internal
		if *intl == ip.Internal {
			// specific NIC requested
			break
		}
	}

	if addr != "" {
		log.Printf("updating existing public ip on %v", *sid)
		st, err = client.Server.UpdatePublicIP(*sid, addr, pubip)
	} else {
		log.Printf("creating public ip on %v. internal: %v", *sid)
		st, err = client.Server.AddPublicIP(*sid, pubip)
	}
	if err != nil {
		log.Panicf("error sending public ip: %v", err)
	}
	log.Printf("polling status on %v", st.ID)
	poll := make(chan *status.Response, 1)
	_ = client.Status.Poll(st.ID, poll)
	status := <-poll
	log.Printf("done. status: %v", status)

	// fetch/print public ips
	svr, _ = client.Server.Get(*sid)
	for _, ip := range svr.Details.IPaddresses {
		log.Printf("IP: %v \t => %v", ip.Internal, ip.Public)
	}
}
