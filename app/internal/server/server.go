package server

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/codecrafters-io/http-server-starter-go/app/internal/middleware"
	"github.com/codecrafters-io/http-server-starter-go/app/internal/network"
	"github.com/codecrafters-io/http-server-starter-go/app/pkg/constants"
	"github.com/codecrafters-io/http-server-starter-go/app/pkg/util"
	"net"
	"strings"
)

func Handle(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)

	for {
		req, err := network.ReadRequest(reader)
		if err != nil {
			return
		}

		if b, err := json.Marshal(req); err == nil {
			fmt.Println(string(b))
		}

		connectionHeader := ""
		if util.ShouldClose(req) {
			connectionHeader = "close"
		}

		strSlice := strings.Split(req.Headers[constants.AcceptEncoding], ",")

		shouldAddEncodingHeader := middleware.ShouldCompress(strSlice)

		dispatchRequest(req, conn, shouldAddEncodingHeader)

		if connectionHeader == "close" {
			return
		}
	}
}
