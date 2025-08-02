package network

import (
	"fmt"
	"github.com/codecrafters-io/http-server-starter-go/app/pkg/constants"
	"net"
	"net/http"
	"strings"
)

type Response struct {
	StatusCode int
	Headers    http.Header
	Body       []byte
}

func NewResponse(statusCode int, body []byte, headers http.Header) *Response {
	return &Response{
		StatusCode: statusCode,
		Headers:    headers,
		Body:       body,
	}
}

func (r *Response) WithHeader(key, value string) *Response {
	r.Headers.Set(key, value)
	return r
}

func (r *Response) toBytes() []byte {
	var b strings.Builder
	statusLine := fmt.Sprintf("HTTP/1.1 %d %s\r\n", r.StatusCode, http.StatusText(r.StatusCode))
	b.WriteString(statusLine)
	for k, vals := range r.Headers {
		for _, v := range vals {
			b.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
		}
	}
	b.WriteString("\r\n")
	if len(r.Body) > 0 {
		b.Write(r.Body)
	}
	return []byte(b.String())
}

func (r *Response) WriteTo(conn net.Conn, shouldClose bool) error {

	if shouldClose {
		r.Headers.Set(constants.Connection, "close")
	}

	fmt.Printf("Response:\n%s\n", string(r.toBytes()))
	_, err := conn.Write(r.toBytes())
	return err
}
