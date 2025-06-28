package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

type Client struct {
	ServerIp   string
	ServerPort int
}

func NewClient(Ip string, Port int) *Client {
	return &Client{
		ServerIp:   Ip,
		ServerPort: Port,
	}
}

func (client *Client) Start() {
	Conn, err := net.Dial("tcp",fmt.Sprintf("%s:%d",client.ServerIp,client.ServerPort))

	//和服务器连接失败
	if err!=nil {
		fmt.Println("与服务器的连接失败")
		return
	}

	defer Conn.Close()

	fmt.Println("成功和服务器建立连接")

	//开启接收和发送协程
	go client.MsgRecv(Conn)
	go client.MsgSend(Conn)

	select{}
}

//信息接收协程
func (client *Client) MsgRecv(Conn net.Conn){
	reader := bufio.NewReader(Conn)
	for {
		msg,err:=reader.ReadString('\n')
		//连接已经失效
		if err!=nil {
			fmt.Println("Connection lost")
			os.Exit(0)
		}
		fmt.Print(msg)
	}
}

//信息发送协程
func (client *Client) MsgSend(Conn net.Conn){
	scanner:=bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		msg:=scanner.Text()
		Conn.Write([]byte(msg+"\n"))
	}
}

