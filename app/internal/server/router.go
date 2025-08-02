package server

import (
	handlers2 "github.com/codecrafters-io/http-server-starter-go/app/internal/handlers"
	"github.com/codecrafters-io/http-server-starter-go/app/internal/network"
	"net"
)

func dispatchRequest(req network.Request, conn net.Conn, shouldAddEncodingHeader bool) {
	switch req.Method {
	case "GET":
		handlers2.HandleGet(req, conn, shouldAddEncodingHeader)
	case "POST":
		handlers2.HandlePost(req, conn)
	default:
		handlers2.HandleMethodNotAllowed(req, conn)
	}
}
