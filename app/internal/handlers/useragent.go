package handlers

import (
	"github.com/codecrafters-io/http-server-starter-go/app/internal/network"
	"github.com/codecrafters-io/http-server-starter-go/app/pkg/constants"
	"github.com/codecrafters-io/http-server-starter-go/app/pkg/util"
	"net"
	"net/http"
	"strconv"
)

func handleUserAgent(request network.Request, connection net.Conn) {
	userAgent := request.Headers[constants.UserAgent]
	body := []byte(userAgent)
	headers := make(http.Header)

	headers.Set(constants.ContentType, "text/plain")
	headers.Set(constants.ContentLength, strconv.Itoa(len(body)))

	response := network.NewResponse(200, body, headers)

	response.WriteTo(connection, util.ShouldClose(request))

}
