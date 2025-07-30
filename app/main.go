package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

var _ = net.Listen
var _ = os.Exit

const CRLF = "\r\n"

func main() {

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	conn, err := l.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}

	fmt.Println("Connection accepted from", conn.RemoteAddr())
	buffer := make([]byte, 1024)
	_, err = conn.Read(buffer)
	if err != nil {
		fmt.Println("Error reading from connection: ", err.Error())
	}
	fmt.Println("Received data:", string(buffer))

	req := string(buffer)
	lines := strings.Split(req, CRLF)
	path := strings.Split(lines[0], " ")[1]
	fmt.Println(path)
	var res string
	if path == "/" {
		res = "HTTP/1.1 200 OK\r\n\r\n"
	} else {
		res = "HTTP/1.1 404 Not Found\r\n\r\n"
	}
	fmt.Println(res)
	conn.Write([]byte(res))

	conn.Close()
	l.Close()

}
