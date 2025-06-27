package main

import "net"

type user struct {
	Name string
	Addr net.Addr
	C    chan string
	Conn net.Conn
	Login bool
}

func NewUser(Name string,Addr net.Addr,Conn net.Conn) *user {
	return &user{
		Name: Name,
		Addr: Addr,
		C: make(chan string),
		Conn: Conn,
		Login: false,
	}
}