package handlers

import (
	"github.com/codecrafters-io/http-server-starter-go/app/internal/network"
	"github.com/codecrafters-io/http-server-starter-go/app/pkg/constants"
	"net"
	"strings"
)

func HandlePost(request network.Request, connection net.Conn) {
	if strings.HasPrefix(request.Path, constants.FilesEndpoint) {
		handleFilePost(request, connection)
		return
	}
	handleNotFound(request, connection)
	return
}
