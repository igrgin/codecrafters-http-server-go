package handlers

import (
	network2 "github.com/codecrafters-io/http-server-starter-go/app/internal/network"
	"net"
	"net/http"
)

func handleDefault(request network2.Request, connection net.Conn) {
	response := network2.NewResponse(http.StatusOK, []byte{}, make(http.Header))
	response.WriteTo(connection)
}
