package main

import "sshproxy/server"

type ProxyEntry struct {
	server.Server
	SocksPort int `json:"socksport"`
	HttpPort  int `json:"httpport"`
}
