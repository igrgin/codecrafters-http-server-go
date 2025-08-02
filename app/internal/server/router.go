package server

import (
	"github.com/codecrafters-io/http-server-starter-go/app/internal/handlers"
	"github.com/codecrafters-io/http-server-starter-go/app/internal/network"
	"net"
)

func dispatchRequest(req network.Request, conn net.Conn, shouldAddEncodingHeader bool) {
	switch req.Method {
	case "GET":
		handlers.HandleGet(req, conn, shouldAddEncodingHeader)
	case "POST":
		handlers.HandlePost(req, conn)
	default:
		handlers.HandleMethodNotAllowed(req, conn)
	}
}
