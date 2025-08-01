package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
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

func gzipBytes(src []byte) ([]byte, error) {
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	_, err := gw.Write(src)
	if err != nil {
		_ = gw.Close()
		return nil, err
	}
	if err := gw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func handle(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)

	for {
		// Request line
		line, err := reader.ReadString('\n')
		if err != nil {
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
				continue
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

		if b, err := json.Marshal(req); err == nil {
			fmt.Println(string(b))
		}

		// Determine if we should signal close
		connectionHeader := ""
		if shouldClose(req) {
			connectionHeader = "close"
		}

		// Determine gzip acceptance
		encodingHeader := ""
		acceptedEncoding := strings.Split(req.Headers["Accept-Encoding"], ",")
		if acceptsGzip(acceptedEncoding) {
			encodingHeader = "Content-Encoding: gzip"
		}

		// Dispatch
		switch req.Method {
		case "GET":
			handleGetWithConnHeader(req, conn, encodingHeader, connectionHeader)
		case "POST":
			fmt.Println("POST")
			handlePostWithConnHeader(req, conn, encodingHeader, connectionHeader)
		default:
			writeResponseWithConnection(conn, "HTTP/1.1 405 Method Not Allowed", nil, nil, connectionHeader)
		}

		if connectionHeader == "close" {
			return
		}
	}
}

func shouldClose(request Request) bool {
	connHdr := request.Headers["Connection"]
	if strings.EqualFold(connHdr, "close") {
		return true
	}
	if request.Protocol == "HTTP/1.0" && !strings.EqualFold(connHdr, "keep-alive") {
		return true
	}
	return false
}

func writeResponse(conn net.Conn, statusLine string, headers map[string]string, body []byte) {
	var sb strings.Builder
	sb.WriteString(statusLine + "\r\n")
	for k, v := range headers {
		sb.WriteString(k + ": " + v + "\r\n")
	}
	sb.WriteString("\r\n")
	conn.Write([]byte(sb.String()))
	if len(body) > 0 {
		conn.Write(body)
	}
}

func writeResponseWithConnection(conn net.Conn, statusLine string, headers map[string]string, body []byte, connectionValue string) {
	if headers == nil {
		headers = make(map[string]string)
	}
	if connectionValue != "" {
		headers["Connection"] = connectionValue
	}
	writeResponse(conn, statusLine, headers, body)
}

func acceptsGzip(acceptedEnc []string) bool {
	for _, part := range acceptedEnc {
		if strings.EqualFold(strings.TrimSpace(part), "gzip") {
			return true
		}
	}
	return false
}

func handlePostWithConnHeader(request Request, connection net.Conn, encodingHeader, connectionHeader string) {
	if strings.HasPrefix(request.Path, "/files/") {
		name := request.Path[len("/files/"):]
		target := *filepath + name
		fmt.Println(target)
		err := os.WriteFile(target, []byte(request.Body), 0644)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error writing file:", err)
			body := []byte(err.Error())
			headers := map[string]string{}
			if acceptsGzip(strings.Split(request.Headers["Accept-Encoding"], ",")) {
				if compressed, err := gzipBytes(body); err == nil {
					body = compressed
					headers["Content-Encoding"] = "gzip"
				} else {
					fmt.Fprintln(os.Stderr, "gzip error:", err)
				}
			}
			headers["Content-Length"] = strconv.Itoa(len(body))
			writeResponseWithConnection(connection, "HTTP/1.1 500 Internal Server Error", headers, body, connectionHeader)
			return
		}
		headers := map[string]string{}
		if acceptsGzip(strings.Split(request.Headers["Accept-Encoding"], ",")) {
			// no body, so just set header if desired (but empty body)
			headers["Content-Encoding"] = "gzip"
		}
		writeResponseWithConnection(connection, "HTTP/1.1 201 Created", headers, nil, connectionHeader)
		return
	}
	writeResponseWithConnection(connection, "HTTP/1.1 404 Not Found", nil, nil, connectionHeader)
}

func handleGetWithConnHeader(request Request, connection net.Conn, encodingHeader, connectionHeader string) {
	switch {
	case request.Path == "/":
		writeResponseWithConnection(connection, "HTTP/1.1 200 OK", nil, nil, connectionHeader)
	case strings.HasPrefix(request.Path, "/echo/"):
		response := strings.TrimPrefix(request.Path, "/echo/")
		body := []byte(response)
		headers := map[string]string{
			"Content-Type": "text/plain",
		}

		if acceptsGzip(strings.Split(request.Headers["Accept-Encoding"], ",")) {
			if compressed, err := gzipBytes(body); err == nil {
				body = compressed
				headers["Content-Encoding"] = "gzip"
			} else {
				fmt.Fprintln(os.Stderr, "gzip error:", err)
			}
		}

		headers["Content-Length"] = strconv.Itoa(len(body))
		fmt.Println(headers)
		fmt.Println(response)
		writeResponseWithConnection(connection, "HTTP/1.1 200 OK", headers, body, connectionHeader)
	case strings.HasPrefix(request.Path, "/user-agent"):
		userAgent := request.Headers["User-Agent"]
		body := []byte(userAgent)
		headers := map[string]string{
			"Content-Type": "text/plain",
		}
		if acceptsGzip(strings.Split(request.Headers["Accept-Encoding"], ",")) {
			if compressed, err := gzipBytes(body); err == nil {
				body = compressed
				headers["Content-Encoding"] = "gzip"
			} else {
				fmt.Fprintln(os.Stderr, "gzip error:", err)
			}
		}
		headers["Content-Length"] = strconv.Itoa(len(body))
		fmt.Println(userAgent)
		writeResponseWithConnection(connection, "HTTP/1.1 200 OK", headers, body, connectionHeader)
	case strings.HasPrefix(request.Path, "/files/"):
		name := request.Path[len("/files/"):]
		target := *filepath + name
		file, err := os.Open(target)
		if err != nil {
			writeResponseWithConnection(connection, "HTTP/1.1 404 Not Found", nil, nil, connectionHeader)
			return
		}
		defer file.Close()
		content, err := io.ReadAll(file)
		if err != nil {
			writeResponseWithConnection(connection, "HTTP/1.1 404 Not Found", nil, nil, connectionHeader)
			return
		}
		body := content
		headers := map[string]string{
			"Content-Type": "application/octet-stream",
		}
		if acceptsGzip(strings.Split(request.Headers["Accept-Encoding"], ",")) {
			if compressed, err := gzipBytes(body); err == nil {
				body = compressed
				headers["Content-Encoding"] = "gzip"
			} else {
				fmt.Fprintln(os.Stderr, "gzip error:", err)
			}
		}
		headers["Content-Length"] = strconv.Itoa(len(body))
		writeResponseWithConnection(connection, "HTTP/1.1 200 OK", headers, body, connectionHeader)
	default:
		writeResponseWithConnection(connection, "HTTP/1.1 404 Not Found", nil, nil, connectionHeader)
	}
}
