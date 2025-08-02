package handlers

import (
	"github.com/codecrafters-io/http-server-starter-go/app/internal/network"
	"github.com/codecrafters-io/http-server-starter-go/app/pkg/util"
	"net"
	"net/http"
)

func handleNotFound(request network.Request, connection net.Conn) {
	response := network.NewResponse(http.StatusNotFound, request.Protocol, []byte{}, make(http.Header))
	response.WriteTo(connection, util.ShouldClose(request))
}
