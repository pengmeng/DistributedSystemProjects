package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
)

const (
	HOST = "localhost"
	TYPE = "tcp"
)

var (
	Clients map[string]int
	Pool    map[int]*net.Conn
	mutex   *sync.Mutex
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Want one and only one argument.")
		os.Exit(1)
	}
	Clients = make(map[string]int)
	Pool = make(map[int]*net.Conn)
	mutex = &sync.Mutex{}
	Clients["next"] = 0
	port := os.Args[1]
	listener, err := net.Listen(TYPE, HOST+":"+port)
	checkErr(err)
	fmt.Println("Listening on port : " + port)
	for {
		conn, err := listener.Accept()
		defer conn.Close()
		checkErr(err)
		go handleRequest(conn)
	}
}

func checkErr(err error) {
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func handleRequest(conn net.Conn) {
	buf := make([]byte, 1024)
	for {
		length, err := conn.Read(buf)
		if err != nil {
			fmt.Println(err.Error())
			conn.Close()
			break
		}
		fmt.Printf(
			"Received: %s from %s\n",
			string(buf[:length]),
			conn.RemoteAddr().String(),
		)
		remote := conn.RemoteAddr().String()
		if _, ok := Clients[remote]; !ok {
			mutex.Lock()
			Clients[remote] = Clients["next"]
			Clients["next"]++
			Pool[Clients[remote]] = &conn
			mutex.Unlock()
		}
		result := strings.TrimSpace(string(buf[:length-1]))
		parts := strings.Split(result, ":")
		if len(parts) == 1 {
			responseAll(conn, parts[0])
		} else {
			router(parts, conn, Clients[remote])
		}
	}
}

func router(parts []string, conn net.Conn, id int) {
	prefix := strings.TrimSpace(parts[0])
	content := strings.TrimSpace(strings.Join(parts[1:], ":"))
	if v := isDigit(prefix); v != -1 {
		responseToId(conn, content, v)
	} else {
		switch prefix {
		case "whoami":
			response(conn, "chitter: "+strconv.Itoa(id))
		case "all":
			responseAll(conn, content)
		default:
			response(conn, "Unknown prefix.")
		}
	}
}

func isDigit(s string) int {
	if v, err := strconv.Atoi(s); err != nil {
		return -1
	} else {
		return v
	}
}

func response(conn net.Conn, content string) {
	fmt.Println("Send ", content)
	conn.Write([]byte(content + "\n"))
}

func responseToId(conn net.Conn, content string, toid int) {
	fromid := Clients[conn.RemoteAddr().String()]
	for id, conn := range Pool {
		if id == toid {
			response(*conn, strconv.Itoa(fromid)+": "+content)
			return
		}
	}
	response(conn, "Id "+strconv.Itoa(toid)+" doesn't exist.")
}

func responseAll(conn net.Conn, content string) {
	from := conn.RemoteAddr().String()
	fromid := Clients[from]
	s := strconv.Itoa(fromid) + ": " + content
	for _, conn := range Pool {
		response(*conn, s)
	}
}
