package handlers

import (
	"fmt"
	"github.com/codecrafters-io/http-server-starter-go/app/internal/middleware"
	network2 "github.com/codecrafters-io/http-server-starter-go/app/internal/network"
	"github.com/codecrafters-io/http-server-starter-go/app/pkg/constants"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func handleEcho(request network2.Request, connection net.Conn, shouldAddEncodingHeader bool) {
	body := []byte(strings.TrimPrefix(request.Path, "/echo/"))
	headers := make(http.Header)

	headers.Set(constants.ContentLength, strconv.Itoa(len(body)))
	if shouldAddEncodingHeader {
		compressed, err := middleware.GzipBytes(body)
		if err == nil {
			body = compressed
			headers.Set(constants.ContentEncoding, "gzip")
		} else {
			fmt.Fprintln(os.Stderr, "gzip error:", err)
		}
	}

	fmt.Println(headers)
	response := network2.NewResponse(200, body, headers)

	response.WriteTo(connection)
}
