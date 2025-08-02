package handlers

import (
	"fmt"
	"github.com/codecrafters-io/http-server-starter-go/app/internal/config"
	"github.com/codecrafters-io/http-server-starter-go/app/internal/network"
	"github.com/codecrafters-io/http-server-starter-go/app/pkg/constants"
	"github.com/codecrafters-io/http-server-starter-go/app/pkg/util"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"
)

func handleFileGet(request network.Request, connection net.Conn) {
	name := request.Path[len(constants.FilesEndpoint):]
	target := config.Instance.Directory + name
	file, err := os.Open(target)
	if err != nil {
		response := network.NewResponse(http.StatusNotFound, []byte{}, make(http.Header))
		response.WriteTo(connection, util.ShouldClose(request))
		return
	}
	defer file.Close()
	content, err := io.ReadAll(file)
	if err != nil {
		response := network.NewResponse(http.StatusNotFound, []byte{}, make(http.Header))
		response.WriteTo(connection, util.ShouldClose(request))
		return
	}
	body := content
	headers := make(http.Header, 1)

	headers.Set(constants.ContentType, "application/octet-stream")

	headers.Set(constants.ContentLength, strconv.Itoa(len(body)))
	response := network.NewResponse(http.StatusOK, body, headers)
	response.WriteTo(connection, util.ShouldClose(request))
}

func handleFilePost(request network.Request, connection net.Conn) {
	name := request.Path[len(constants.FilesEndpoint):]
	header := make(http.Header)
	target := config.Instance.Directory + name
	fmt.Println("TARGET: " + target)
	err := os.WriteFile(target, []byte(request.Body), 0644)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error writing file:", err)
		response := network.NewResponse(http.StatusNotFound, []byte{}, header)
		response.WriteTo(connection, util.ShouldClose(request))
	}

	response := network.NewResponse(http.StatusCreated, []byte{}, header)
	response.WriteTo(connection, util.ShouldClose(request))
}
