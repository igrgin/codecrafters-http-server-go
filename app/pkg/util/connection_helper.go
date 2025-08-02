package util

import (
	"github.com/codecrafters-io/http-server-starter-go/app/internal/network"
	"github.com/codecrafters-io/http-server-starter-go/app/pkg/constants"
	"strings"
)

func ShouldClose(request network.Request) bool {
	connHdr := request.Headers[constants.Connection]
	if strings.EqualFold(connHdr, "close") {
		return true
	}

	return false
}
