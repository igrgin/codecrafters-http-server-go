package network

import (
	"bufio"
	"fmt"
	"github.com/codecrafters-io/http-server-starter-go/app/pkg/constants"
	"io"
	"os"
	"strconv"
	"strings"
)

type Request struct {
	Method   string
	Path     string
	Protocol string
	Headers  map[string]string
	Body     string
}

func ReadRequest(reader *bufio.Reader) (Request, error) {
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
	cl, doesContentLengthExists := req.Headers[constants.ContentLength]
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
