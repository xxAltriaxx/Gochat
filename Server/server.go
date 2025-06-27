package main

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
)

type Server struct {
	Ip      string
	Port    int
	Users   map[net.Conn]user
	PubChan chan string
	mu      sync.RWMutex
}

func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:      ip,
		Port:    port,
		Users:   make(map[net.Conn]user),
		PubChan: make(chan string),
	}
	return server
}

func (server *Server) Start() {
	//socket listen
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", server.Port))
	if err != nil {
		fmt.Println("Listen Error!")
		return
	} 

	fmt.Println("服务器成功启动")

	//close socket
	defer listener.Close()

	//start broadcast
	go server.broadcast()

	for {
		//accept
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("accept Error!")
			continue
		}

		//handle
		go server.Handle(conn)
	}
}

// 处理连接
func (server *Server) Handle(Conn net.Conn) {
	reader := bufio.NewReader(Conn)
	for {
		msg, err := reader.ReadString('\n')
		if err != nil {
			server.CloseConn(Conn)
		}

		switch {
		case !server.Users[Conn].Login:
			switch {
			//提供帮助/help
			case strings.HasPrefix(msg, "/help"):
				{
					bytes := []byte("* /help ->Get some help\n" +
						"* /joinAs [name] ->Join the server and set username\n" +
						"* /msg [userame] [message] ->Private Chat\n" +
						"* /users ->Check users\n" +
						"* /quit ->Quit\n")
					Conn.Write(bytes)
				}

			//加入服务器/joinAs
			case strings.HasPrefix(msg, "/joinAs "):
				{
					//检查指令格式
					strs := strings.Split(msg, " ")
					if len(strs) != 2 {
						Conn.Write([]byte("Wrong command format!\n"))
						continue
					}

					//正式将用户添加进入用户列表
					server.mu.Lock()
					user:=NewUser(strs[1],Conn.RemoteAddr(),Conn)
					server.Users[Conn]=*user
					server.mu.Unlock()
				}
			}
		default:
			{
				switch {
				//私聊/msg
				case strings.HasPrefix(msg,"/msg ") :{
					//检查指令格式
					strs := strings.Split(msg, " ")
					if len(strs) != 3 {
						Conn.Write([]byte("Wrong command format!\n"))
						continue
					}

					//将消息放入匹配用户通道
					for _,user:=range server.Users{
						if user.Name==strs[1] {
							user.C <- "From "+user.Name+" privately: "+strs[2]
							continue
						}
					}
					
					//该用户不在线
					Conn.Write([]byte("User isn't online\n"))
				}

				//查看在线用户/users
				case strings.EqualFold(msg,"/users") :{
					Conn.Write([]byte(strconv.Itoa(len(server.Users))+" user online currently\n"))
					for _,user:=range server.Users {
						Conn.Write([]byte(user.Name+"\n"))
					}
				}

				//退出
				case strings.EqualFold(msg,"/quit") :{
					server.CloseConn(Conn)
				}

				//发送广播消息
				default :{
					server.PubChan <- msg
				}
				}
			}

		}

		//查看用户通道内是否有他人给该用户发送的消息
		if len(server.Users[Conn].C)>0 {
			Conn.Write([]byte(<-server.Users[Conn].C))
		}
	}
}

// 关闭连接
func (s *Server) CloseConn(conn net.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if user, exists := s.Users[conn]; exists {
		fmt.Fprintf(conn, "连接关闭!\n")
		delete(s.Users, conn)
		s.PubChan <- fmt.Sprintf("系统: %s 离开聊天室", user.Name)
		conn.Close()
	}
}

// 广播信息
func (server *Server) broadcast() {
	for {
		msg := <-server.PubChan
		server.mu.Lock()
		for _, user := range server.Users {
			user.Conn.Write([]byte(msg))
		}
		server.mu.Unlock()
	}
}
