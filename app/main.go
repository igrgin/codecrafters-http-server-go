package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
)

// Ensures gofmt doesn't remove the "net" and "os" imports above (feel free to remove this!)
var (
	_        = net.Listen
	_        = os.Exit
	filepath = flag.String("directory", "", "directory to serve files")
)

func main() {
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}

	defer l.Close()
	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		flag.Parse()
		go handle(conn)
	}
}

type Request struct {
	Method   string
	Path     string
	Protocol string
	Headers  map[string]string
	Body     string
}

func handle(connection net.Conn) net.Conn {
	reader := bufio.NewReader(connection)

	response, _ := reader.ReadString('\n')

	requestParts := strings.Split(response, " ")

	request := Request{
		Method:   requestParts[0],
		Path:     requestParts[1],
		Protocol: strings.TrimSuffix(requestParts[2], "\r\n"),
		Headers:  make(map[string]string),
		Body:     "",
	}

	for {
		header, _ := reader.ReadString('\n')
		if header == "\r\n" {
			break
		}
		headerParts := strings.Split(header, ":")
		request.Headers[headerParts[0]] = strings.TrimSpace(strings.Join(headerParts[1:], ""))
	}
	contentSize, err := strconv.Atoi(request.Headers["Content-Length"])
	fmt.Println(err)
	if err == nil {
		content := make([]byte, contentSize)
		size, err := io.ReadFull(reader, content)
		if size != contentSize {
			log.Fatal("dead")
		}
		if err != nil {
			fmt.Println(err)
		} else {
			request.Body = string(content)
		}
	}
	s, err := json.Marshal(request)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println(string(s))
	connection = handleRequest(request, connection)

	return connection
}

func handleRequest(request Request, connection net.Conn) net.Conn {
	defer connection.Close()

	switch {
	case request.Method == "GET":
		{
			handleGet(request, connection)
		}
	case request.Method == "POST":
		{
			fmt.Println("POST")
			handlePost(request, connection)
		}
	default:
		{
			errors.New("Unhandled Method")
		}
	}
	return connection
}

func handlePost(request Request, connection net.Conn) net.Conn {
	switch {
	case strings.HasPrefix(request.Path, "/files/"):
		{
			d1 := []byte(request.Body)
			name := request.Path[len("/files/"):]
			fmt.Println(*filepath + name)
			err := os.WriteFile(*filepath+name, d1, 0644)

			if err != nil {
				fmt.Println("Error", err)
				connection.Write([]byte(err.Error()))
				return connection
			}

			connection.Write([]byte("HTTP/1.1 201 Created\r\n\r\n"))
		}
	}
	return connection
}

func handleGet(request Request, connection net.Conn) net.Conn {
	switch {
	case request.Path == "/":
		{
			connection.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
		}
	case strings.HasPrefix(request.Path, "/echo/"):
		{
			var sb strings.Builder
			response, _ := strings.CutPrefix(request.Path, "/echo/")
			sb.WriteString(response)
			strResponse := fmt.Sprintf("HTTP/1.1 200 OK\r\n"+
				"Content-Type: text/plain\r\n"+
				"Content-Length: %d\r\n"+
				"Connection: close\r\n\r\n"+
				"%s",
				len(sb.String()),
				sb.String())
			fmt.Println(sb.String())
			fmt.Printf("Response body: [% x]\n", []byte(sb.String()))
			connection.Write([]byte(strResponse))
		}
	case strings.HasPrefix(request.Path, "/user-agent"):
		{
			userAgent := request.Headers["User-Agent"]
			strResponse := fmt.Sprintf("HTTP/1.1 200 OK\r\n"+
				"Content-Type: text/plain\r\n"+
				"Content-Length: %d\r\n"+
				"Connection: close\r\n\r\n"+
				"%s",
				len(userAgent),
				userAgent)
			fmt.Println(userAgent)
			fmt.Printf("Response body: [% x]\n", []byte(userAgent))
			connection.Write([]byte(strResponse))
		}
	case strings.HasPrefix(request.Path, "/files/"):
		{
			name := request.Path[len("/files/"):]
			file, err := os.Open(*filepath + name)
			if err != nil {
				connection.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
				return connection
			}
			defer file.Close()

			content, err := io.ReadAll(file)
			if err != nil {
				connection.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
				return connection
			}

			strResponse := fmt.Sprintf("HTTP/1.1 200 OK\r\n"+
				"Content-Type: application/octet-stream\r\n"+
				"Content-Length: %d\r\n"+
				"Connection: close\r\n\r\n"+
				"%s", len(content), content)

			connection.Write([]byte(strResponse))
		}
	default:
		{
			connection.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
		}

	}

	return connection
}
