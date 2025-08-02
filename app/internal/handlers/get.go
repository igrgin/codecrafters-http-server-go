package handlers

import (
	"github.com/codecrafters-io/http-server-starter-go/app/internal/network"
	"github.com/codecrafters-io/http-server-starter-go/app/pkg/constants"
	"net"
	"strings"
)

func HandleGet(request network.Request, connection net.Conn, shouldAddEncodingHeader bool) {
	switch {
	case request.Path == constants.DefaultEndpoint:
		handleDefault(request, connection)
	case strings.HasPrefix(request.Path, constants.EchoEndpoint):
		handleEcho(request, connection, shouldAddEncodingHeader)
	case strings.HasPrefix(request.Path, constants.UserAgentEndpoint):
		handleUserAgent(request, connection)
	case strings.HasPrefix(request.Path, constants.FilesEndpoint):
		handleFileGet(request, connection)
	default:
		handleNotFound(request, connection)
	}
}
