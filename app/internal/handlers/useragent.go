package handlers

import (
	"fmt"
	network2 "github.com/codecrafters-io/http-server-starter-go/app/internal/network"
	"github.com/codecrafters-io/http-server-starter-go/app/pkg/constants"
	"net"
	"net/http"
	"strconv"
)

func handleUserAgent(request network2.Request, connection net.Conn) {
	userAgent := request.Headers["User-Agent"]
	body := []byte(userAgent)
	headers := make(http.Header)

	headers.Set(constants.ContentLength, strconv.Itoa(len(body)))

	fmt.Println(userAgent)
	response := network2.NewResponse(200, body, headers)

	response.WriteTo(connection)

}
