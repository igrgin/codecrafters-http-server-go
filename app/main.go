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

const FILES_ENDPOINT_PREFIX = "/files/"

const ACCEPT_ENCODING = "Accept-Encoding"
const CONTENT_ENCODING = "Content-Encoding"
const CONTENT_TYPE = "Content-Type"
const CONTENT_LENGTH = "Content-Length"

const STATUS_OK = "HTTP/1.1 200 OK"
const STATUS_NOT_FOUND = "HTTP/1.1 404 Not Found"

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
		req, err := readRequest(reader)
		if err != nil {
			return
		}

		if b, err := json.Marshal(req); err == nil {
			fmt.Println(string(b))
		}

		connectionHeader := ""
		if shouldClose(req) {
			connectionHeader = "close"
		}

		shouldAddEncodingHeader := acceptsGzip(strings.Split(req.Headers[ACCEPT_ENCODING], ","))

		dispatchRequest(req, conn, shouldAddEncodingHeader, connectionHeader)

		if connectionHeader == "close" {
			return
		}
	}
}

func readRequest(reader *bufio.Reader) (Request, error) {
	// Request line
	line, err := reader.ReadString('\n')
	if err != nil {
		return Request{}, err
	}
	parts := strings.SplitN(strings.TrimRight(line, "\r\n"), " ", 3)
	if len(parts) < 3 {
		fmt.Fprintln(os.Stderr, "Malformed request line:", line)
		return Request{}, fmt.Errorf("malformed request line")
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
			return Request{}, err
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

	if err := readRequestBody(reader, &req); err != nil {
		fmt.Fprintln(os.Stderr, "Error reading body:", err)
	}
	return req, nil
}

func readRequestBody(reader *bufio.Reader, req *Request) error {
	cl, doesContentLengthExists := req.Headers[CONTENT_LENGTH]
	if !doesContentLengthExists {
		return nil
	}
	contentSize, err := strconv.Atoi(cl)
	if err != nil || contentSize <= 0 {
		return err
	}
	content := make([]byte, contentSize)
	n, err := io.ReadFull(reader, content)
	if err != nil {
		return err
	}
	if n != contentSize {
		return fmt.Errorf("unexpected body size: %d, expected %d", n, contentSize)
	}
	req.Body = string(content)
	return nil
}

func dispatchRequest(req Request, conn net.Conn, shouldAddEncodingHeader bool, connectionHeader string) {
	switch req.Method {
	case "GET":
		handleGet(req, conn, shouldAddEncodingHeader, connectionHeader)
	case "POST":
		fmt.Println("POST")
		handlePost(req, conn, shouldAddEncodingHeader, connectionHeader)
	default:
		writeResponseWithConnection(conn, "HTTP/1.1 405 Method Not Allowed", nil, nil, connectionHeader)
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
	_, err := conn.Write([]byte(sb.String()))
	if err != nil {
		return
	}
	if len(body) > 0 {
		_, err := conn.Write(body)
		if err != nil {
			return
		}
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

func handlePost(request Request, connection net.Conn, shouldAddEncodingHeader bool, connectionHeader string) {
	if strings.HasPrefix(request.Path, FILES_ENDPOINT_PREFIX) {
		name := request.Path[len(FILES_ENDPOINT_PREFIX):]
		target := *filepath + name
		fmt.Println(target)
		err := os.WriteFile(target, []byte(request.Body), 0644)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error writing file:", err)
			body := []byte(err.Error())
			headers := map[string]string{}
			if shouldAddEncodingHeader {
				if compressed, err := gzipBytes(body); err == nil {
					body = compressed
					headers[CONTENT_ENCODING] = "gzip"
				} else {
					fmt.Fprintln(os.Stderr, "gzip error:", err)
				}
			}
			headers[CONTENT_LENGTH] = strconv.Itoa(len(body))
			writeResponseWithConnection(connection, "HTTP/1.1 500 Internal Server Error", headers, body, connectionHeader)
			return
		}
		headers := map[string]string{}
		if shouldAddEncodingHeader {
			// no body, so just set header if desired (but empty body)
			headers[CONTENT_ENCODING] = "gzip"
		}
		writeResponseWithConnection(connection, "HTTP/1.1 201 Created", headers, nil, connectionHeader)
		return
	}
	writeResponseWithConnection(connection, STATUS_NOT_FOUND, nil, nil, connectionHeader)
}

func handleGet(request Request, connection net.Conn, addEncodingHeader bool, connectionHeader string) {
	switch {
	case request.Path == "/":
		writeResponseWithConnection(connection, STATUS_OK, nil, nil, connectionHeader)
	case strings.HasPrefix(request.Path, "/echo/"):
		handleEcho(request, connection, addEncodingHeader, connectionHeader)
	case strings.HasPrefix(request.Path, "/user-agent"):
		handleUserAgent(request, connection, connectionHeader)
	case strings.HasPrefix(request.Path, FILES_ENDPOINT_PREFIX):
		handleFileGet(request, connection, connectionHeader)
	default:
		writeResponseWithConnection(connection, STATUS_NOT_FOUND, nil, nil, connectionHeader)
	}
}

func handleEcho(request Request, connection net.Conn, addEncodingHeader bool, connectionHeader string) {
	response := strings.TrimPrefix(request.Path, "/echo/")
	body := []byte(response)
	headers := map[string]string{
		CONTENT_TYPE: "text/plain",
	}
	if addEncodingHeader {
		if compressed, err := gzipBytes(body); err == nil {
			body = compressed
			headers[CONTENT_ENCODING] = "gzip"
		} else {
			fmt.Fprintln(os.Stderr, "gzip error:", err)
		}
	}
	headers[CONTENT_LENGTH] = strconv.Itoa(len(body))
	fmt.Println(headers)
	fmt.Println(response)
	writeResponseWithConnection(connection, STATUS_OK, headers, body, connectionHeader)
}

func handleUserAgent(request Request, connection net.Conn, connectionHeader string) {
	userAgent := request.Headers["User-Agent"]
	body := []byte(userAgent)
	headers := map[string]string{
		CONTENT_TYPE: "text/plain",
	}
	if acceptsGzip(strings.Split(request.Headers[ACCEPT_ENCODING], ",")) {
		if compressed, err := gzipBytes(body); err == nil {
			body = compressed
			headers[CONTENT_ENCODING] = "gzip"
		} else {
			fmt.Fprintln(os.Stderr, "gzip error:", err)
		}
	}
	headers[CONTENT_LENGTH] = strconv.Itoa(len(body))
	fmt.Println(userAgent)
	writeResponseWithConnection(connection, STATUS_OK, headers, body, connectionHeader)
}

func handleFileGet(request Request, connection net.Conn, connectionHeader string) {
	name := request.Path[len(FILES_ENDPOINT_PREFIX):]
	target := *filepath + name
	file, err := os.Open(target)
	if err != nil {
		writeResponseWithConnection(connection, STATUS_NOT_FOUND, nil, nil, connectionHeader)
		return
	}
	defer file.Close()
	content, err := io.ReadAll(file)
	if err != nil {
		writeResponseWithConnection(connection, STATUS_NOT_FOUND, nil, nil, connectionHeader)
		return
	}
	body := content
	headers := map[string]string{
		CONTENT_TYPE: "application/octet-stream",
	}
	if acceptsGzip(strings.Split(request.Headers[ACCEPT_ENCODING], ",")) {
		if compressed, err := gzipBytes(body); err == nil {
			body = compressed
			headers[CONTENT_ENCODING] = "gzip"
		} else {
			fmt.Fprintln(os.Stderr, "gzip error:", err)
		}
	}
	headers[CONTENT_LENGTH] = strconv.Itoa(len(body))
	writeResponseWithConnection(connection, STATUS_OK, headers, body, connectionHeader)
}
