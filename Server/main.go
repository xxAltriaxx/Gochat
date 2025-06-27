package main

func main(){
	server:=NewServer("127.0.0.1",1145)
	server.Start()
}