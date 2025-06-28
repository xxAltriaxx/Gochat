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
		//处理私密消息
		go server.PrivateMsg(conn)
	}
}

// 处理连接
func (server *Server) Handle(Conn net.Conn) {
	reader := bufio.NewReader(Conn)
	for {
		msg, err := reader.ReadString('\n')
		msg = msg[0 : len(msg)-1]

		if err != nil {
			server.CloseConn(Conn)
		}

		fmt.Println("Receive msg: " + msg)

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
					user := NewUser(strs[1], Conn.RemoteAddr(), Conn)
					user.Login = true
					server.Users[Conn] = *user
					server.mu.Unlock()

					server.PubChan <- strs[1] + " joined the server\n"
				}
				//未登录无法使用其他功能
			default:
				{
					Conn.Write([]byte("Use /joinAs [name] First!\n"))
				}
			}
		default:
			{
				switch {
				//私聊/msg
				case strings.HasPrefix(msg, "/msg "):
					{
						//检查指令格式
						strs := strings.Split(msg, " ")
						if len(strs) != 3 {
							Conn.Write([]byte("Wrong command format!\n"))
							continue
						}

						sended := false

						//将消息放入匹配用户通道
						server.mu.Lock()
						for Conn, user := range server.Users {
							if user.Name == strs[1] {
								Conn.Write([]byte("From " + user.Name + " privately: " + strs[2] + "\n"))
								sended = true
								break
							}
						}
						server.mu.Unlock()

						//该用户不在线
						if !sended {
							Conn.Write([]byte("User isn't online\n"))
						}
					}

				//查看在线用户/users
				case msg == "/users":
					{
						Conn.Write([]byte(strconv.Itoa(len(server.Users)) + " user online currently\n"))
						for _, user := range server.Users {
							Conn.Write([]byte(user.Name + "\n"))
						}
					}

				//退出
				case msg == "/quit":
					{
						server.CloseConn(Conn)
						return
					}

				//发送广播消息
				default:
					{
						server.PubChan <- server.Users[Conn].Name + " say: " + msg + "\n"
					}
				}
			}
		}
	}
}

// 处理私密消息
func (server *Server) PrivateMsg(Conn net.Conn) {
	server.mu.RLock()
	user, exists := server.Users[Conn]
	server.mu.RUnlock()

	if !exists {
		return
	}

	for msg := range user.C {
		// 检查连接状态
		if _, err := user.Conn.Write([]byte(msg)); err != nil {
			return
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
		s.PubChan <- fmt.Sprintf("%s 离开服务器", user.Name)
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
