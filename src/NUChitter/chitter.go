package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
)

const (
	HOST = "localhost"
	TYPE = "tcp"
)

type Client struct {
	Id   int
	Conn *net.Conn
}

type Message struct {
	Src  int
	Des  int
	Body string
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Want one and only one argument.")
		os.Exit(1)
	}
	port := os.Args[1]
	listener, err := net.Listen(TYPE, HOST+":"+port)
	checkErr(err)
	defer listener.Close()
	fmt.Println("Listening on port : " + port)
	newid := 1
	clientCh := make(chan *Client)
	msgCh := make(chan *Message)
	go handleMsg(clientCh, msgCh)
	for {
		conn, err := listener.Accept()
		checkErr(err)
		defer conn.Close()
		client := &Client{Id: newid, Conn: &conn}
		go handleConn(client, clientCh, msgCh)
		newid++
	}
}

func checkErr(err error) {
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func handleMsg(clientCh chan *Client, msgCh chan *Message) {
	pool := make(map[int]*net.Conn)
	for {
		select {
		case client := <-clientCh:
			if _, ok := pool[client.Id]; ok {
				delete(pool, client.Id)
				fmt.Printf("Id %d left. %d client(s) active.\n", client.Id, len(pool))
			} else {
				pool[client.Id] = client.Conn
				fmt.Printf("Id %d join. %d client(s) active.\n", client.Id, len(pool))
			}
		case message := <-msgCh:
			if message.Des == 0 {
				for _, conn := range pool {
					response(*conn, message.Body)
				}
			} else {
				if conn, ok := pool[message.Des]; ok {
					response(*conn, message.Body)
				} else {
					response(*pool[message.Src], "Unknown client Id.")
				}
			}
		}
	}
}

func handleConn(client *Client, clientCh chan *Client, msgCh chan *Message) {
	buf := make([]byte, 1024)
	clientCh <- client
	for {
		conn := *client.Conn
		length, err := conn.Read(buf)
		if err != nil {
			conn.Close()
			break
		}
		fmt.Printf(
			"Received: %s from %s\n",
			string(buf[:length-1]),
			conn.RemoteAddr().String(),
		)
		message := parseMsg(client.Id, string(buf[:length-1]))
		msgCh <- message
	}
	clientCh <- client
}

func parseMsg(src int, body string) *Message {
	message := &Message{
		Src: src,
		Des: 0,
	}
	parts := strings.SplitN(strings.TrimSpace(body), ":", 2)
	if len(parts) == 1 {
		message.Body = strconv.Itoa(src) + ": " + parts[0]
	} else {
		prefix := strings.TrimSpace(parts[0])
		body := strings.TrimSpace(parts[1])
		if v, err := strconv.Atoi(prefix); err == nil {
			message.Body = strconv.Itoa(src) + ": " + body
			message.Des = v
		} else {
			switch prefix {
			case "whoami":
				message.Body = "chitter: " + strconv.Itoa(src)
				message.Des = src
			case "all":
				message.Body = strconv.Itoa(src) + ": " + body
			default:
				message.Body = "Unknown prefix."
				message.Des = src
			}
		}
	}
	return message
}

func response(conn net.Conn, content string) {
	fmt.Println("Send ", content)
	conn.Write([]byte(content + "\n"))
}
