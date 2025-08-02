package main

import (
	"flag"
	"fmt"
	"github.com/codecrafters-io/http-server-starter-go/app/internal/config"
	"github.com/codecrafters-io/http-server-starter-go/app/internal/server"
	"net"
	"os"
)

var filepath = flag.String("directory", "", "directory to serve files")

func main() {
	flag.Parse()

	config.Instance.Directory = *filepath

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to bind to port 4221:", err)
		os.Exit(1)
	}
	defer l.Close()

	for {
		conn, err := l.Accept()
		defer conn.Close()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error accepting connection:", err)
			os.Exit(1)
		}
		go server.Handle(conn)
	}
}
