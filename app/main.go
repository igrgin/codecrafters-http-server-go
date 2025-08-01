package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
)

var filepath = flag.String("directory", "", "directory to serve files")

type Request struct {
	Method   string
	Path     string
	Protocol string
	Headers  map[string]string
	Body     string
}

func main() {
	flag.Parse()

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to bind to port 4221:", err)
		os.Exit(1)
	}
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error accepting connection:", err)
			os.Exit(1)
		}
		go handle(conn)
	}
}

func handle(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)

	// Request line
	line, err := reader.ReadString('\n')
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error reading request line:", err)
		return
	}
	parts := strings.SplitN(strings.TrimRight(line, "\r\n"), " ", 3)
	if len(parts) < 3 {
		fmt.Fprintln(os.Stderr, "Malformed request line:", line)
		return
	}
	req := Request{
		Method:   parts[0],
		Path:     parts[1],
		Protocol: parts[2],
		Headers:  make(map[string]string),
	}

	// Headers
	for {
		headerLine, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error reading header:", err)
			return
		}
		if headerLine == "\r\n" {
			break
		}
		colon := strings.Index(headerLine, ":")
		if colon == -1 {
			continue // skip malformed header
		}
		key := headerLine[:colon]
		value := strings.TrimSpace(headerLine[colon+1:])
		req.Headers[key] = value
	}

	// Body, if any
	if cl, ok := req.Headers["Content-Length"]; ok {
		contentSize, err := strconv.Atoi(cl)
		if err == nil && contentSize > 0 {
			content := make([]byte, contentSize)
			n, err := io.ReadFull(reader, content)
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error reading body:", err)
			} else if n != contentSize {
				fmt.Fprintln(os.Stderr, "Unexpected body size:", n, "expected", contentSize)
			} else {
				req.Body = string(content)
			}
		}
	}

	// Debug dump of request
	if b, err := json.Marshal(req); err == nil {
		fmt.Println(string(b))
	}

	handleRequest(req, conn)
}

func writeResponse(conn net.Conn, statusLine string, headers map[string]string, body []byte) {
	var sb strings.Builder
	sb.WriteString(statusLine + "\r\n")
	for k, v := range headers {
		sb.WriteString(k + ": " + v + "\r\n")
	}
	sb.WriteString("Connection: close\r\n")
	sb.WriteString("\r\n")
	conn.Write([]byte(sb.String()))
	if len(body) > 0 {
		conn.Write(body)
	}
}

func acceptsGzip(acceptedEnc []string) bool {
	for _, part := range acceptedEnc {
		if strings.EqualFold(strings.TrimSpace(part), "gzip") {
			return true
		}
	}
	return false
}

func handleRequest(request Request, connection net.Conn) {
	// connection is closed by caller defer
	encodingHeader := ""
	acceptedEncoding := strings.Split(request.Headers["Accept-Encoding"], ",")
	if acceptsGzip(acceptedEncoding) {
		encodingHeader = "Content-Encoding: gzip"
	}

	switch request.Method {
	case "GET":
		handleGet(request, connection, encodingHeader)
	case "POST":
		fmt.Println("POST")
		handlePost(request, connection, encodingHeader)
	default:
		// Unhandled method
		writeResponse(connection, "HTTP/1.1 405 Method Not Allowed", map[string]string{}, nil)
	}
}

func handlePost(request Request, connection net.Conn, encodingHeader string) {
	if strings.HasPrefix(request.Path, "/files/") {
		name := request.Path[len("/files/"):]
		target := *filepath + name
		fmt.Println(target)
		err := os.WriteFile(target, []byte(request.Body), 0644)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error writing file:", err)
			writeResponse(connection, "HTTP/1.1 500 Internal Server Error", map[string]string{}, []byte(err.Error()))
			return
		}
		headers := map[string]string{}
		if encodingHeader != "" {
			headers["Content-Encoding"] = "gzip"
		}
		writeResponse(connection, "HTTP/1.1 201 Created", headers, nil)
		return
	}
	// Fallback for unsupported POST paths
	writeResponse(connection, "HTTP/1.1 404 Not Found", map[string]string{}, nil)
}

func handleGet(request Request, connection net.Conn, encodingHeader string) {
	switch {
	case request.Path == "/":
		writeResponse(connection, "HTTP/1.1 200 OK", map[string]string{}, nil)
	case strings.HasPrefix(request.Path, "/echo/"):
		response := strings.TrimPrefix(request.Path, "/echo/")
		headers := map[string]string{
			"Content-Type":   "text/plain",
			"Content-Length": strconv.Itoa(len(response)),
		}

		if encodingHeader != "" {
			headers["Content-Encoding"] = "gzip"
		}

		fmt.Println(headers)

		fmt.Println(response)
		writeResponse(connection, "HTTP/1.1 200 OK", headers, []byte(response))
	case strings.HasPrefix(request.Path, "/user-agent"):
		userAgent := request.Headers["User-Agent"]
		headers := map[string]string{
			"Content-Type":   "text/plain",
			"Content-Length": strconv.Itoa(len(userAgent)),
		}
		if encodingHeader != "" {
			headers["Content-Encoding"] = "gzip"
		}
		fmt.Println(userAgent)
		writeResponse(connection, "HTTP/1.1 200 OK", headers, []byte(userAgent))
	case strings.HasPrefix(request.Path, "/files/"):
		name := request.Path[len("/files/"):]
		target := *filepath + name
		file, err := os.Open(target)
		if err != nil {
			writeResponse(connection, "HTTP/1.1 404 Not Found", map[string]string{}, nil)
			return
		}
		defer file.Close()
		content, err := io.ReadAll(file)
		if err != nil {
			writeResponse(connection, "HTTP/1.1 404 Not Found", map[string]string{}, nil)
			return
		}
		headers := map[string]string{
			"Content-Type":   "application/octet-stream",
			"Content-Length": strconv.Itoa(len(content)),
		}
		if encodingHeader != "" {
			headers["Content-Encoding"] = "gzip"
		}
		writeResponse(connection, "HTTP/1.1 200 OK", headers, content)
	default:
		writeResponse(connection, "HTTP/1.1 404 Not Found", map[string]string{}, nil)
	}
}
