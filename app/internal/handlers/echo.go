package handlers

import (
	"fmt"
	"github.com/codecrafters-io/http-server-starter-go/app/internal/middleware"
	"github.com/codecrafters-io/http-server-starter-go/app/internal/network"
	"github.com/codecrafters-io/http-server-starter-go/app/pkg/constants"
	"github.com/codecrafters-io/http-server-starter-go/app/pkg/util"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func handleEcho(request network.Request, connection net.Conn, shouldAddEncodingHeader bool) {
	body := []byte(strings.TrimPrefix(request.Path, "/echo/"))
	headers := make(http.Header)
	if shouldAddEncodingHeader {
		headers.Set(constants.ContentType, "text/plain")
		compressed, err := middleware.GzipBytes(body)
		if err == nil {
			body = compressed
			headers.Set(constants.ContentEncoding, "gzip")
			headers.Set(constants.ContentLength, strconv.Itoa(len(body)))
		} else {
			fmt.Fprintln(os.Stderr, "gzip error:", err)
		}
	} else {
		headers.Set(constants.ContentLength, strconv.Itoa(len(body)))
		headers.Set(constants.ContentType, "text/plain")
	}

	fmt.Println(headers)
	response := network.NewResponse(200, body, headers)

	response.WriteTo(connection, util.ShouldClose(request))
}
