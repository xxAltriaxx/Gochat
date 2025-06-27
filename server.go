package main

import (
	"fmt"
	"net"
)

type Server struct {
	Ip   string
	Port int
}

func NewServer(ip string, port int) *Server {
	server := &Server{ip, port}
	return server
}

func (server *Server) Start() {
	//socket listen
	listener,err:= net.Listen("tcp",fmt.Sprintf("%s:%d",server.Ip,server.Port))
	if err!=nil {
		fmt.Println("Listen Error!")
		return
	}

	//close socket
	defer listener.Close()

	for {
		//accept
		conn,err:=listener.Accept()
		if err!=nil {
			fmt.Println("accept Error!")
			continue
		}

		//handle
		go server.Handle(conn)
	}
}

func (server *Server) Handle(conn net.Conn) {
	//处理连接
}