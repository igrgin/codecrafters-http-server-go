package handlers

import (
	"fmt"
	"github.com/codecrafters-io/http-server-starter-go/app/internal/config"
	network2 "github.com/codecrafters-io/http-server-starter-go/app/internal/network"
	constants2 "github.com/codecrafters-io/http-server-starter-go/app/pkg/constants"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"
)

func handleFileGet(request network2.Request, connection net.Conn) {
	name := request.Path[len(constants2.FilesEndpoint):]
	target := config.Configuration.GetDirectory() + name
	file, err := os.Open(target)
	if err != nil {
		response := network2.NewResponse(http.StatusNotFound, []byte{}, make(http.Header))
		response.WriteTo(connection)
		return
	}
	defer file.Close()
	content, err := io.ReadAll(file)
	if err != nil {
		response := network2.NewResponse(http.StatusNotFound, []byte{}, make(http.Header))
		response.WriteTo(connection)
		return
	}
	body := content
	headers := make(http.Header, 1)

	headers.Set(constants2.ContentType, "application/octet-stream")

	headers.Set(constants2.ContentLength, strconv.Itoa(len(body)))
	response := network2.NewResponse(http.StatusOK, body, headers)
	response.WriteTo(connection)
}

func handleFilePost(request network2.Request, connection net.Conn) {
	name := request.Path[len(constants2.FilesEndpoint):]
	header := make(http.Header)
	header.Set(constants2.ContentType, "text/plain")
	target := config.Configuration.GetDirectory() + name
	fmt.Println(target)
	err := os.WriteFile(target, []byte(request.Body), 0644)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error writing file:", err)
		response := network2.NewResponse(http.StatusNotFound, []byte{}, header)
		response.WriteTo(connection)
	}

	response := network2.NewResponse(http.StatusCreated, []byte{}, header)
	response.WriteTo(connection)
}
