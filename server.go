package main

import (
	"net"
)

type Server struct {
	*net.TCPListener
}

func NewServer(addr_str string) (server *Server, err error) {
	var (
		addr *net.TCPAddr
	)

	server = &Server{}

	if addr, err = net.ResolveTCPAddr("tcp", addr_str); err != nil {
		return
	}

	if server.TCPListener, err = net.ListenTCP("tcp", addr); err != nil {
		return
	}

	return
}
