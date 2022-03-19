package http

import (
	"bufio"
	"bytes"
	"fmt"
	"net/http"
	"net/url"
)

func Send(destinationAddress string, msg []byte) ([]byte, error) {
	client := http.Client{}

	br := bufio.NewReader(bytes.NewReader(msg))
	request, err := http.ReadRequest(br)
	if err != nil {
		return nil, err
	}
	if !request.URL.IsAbs() {
		request.RequestURI = ""
		request.URL, _ = url.Parse(fmt.Sprintf("%s%s", destinationAddress, request.URL.Path))
	} else {
		request.URL.Host = destinationAddress
	}
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	buf := bytes.Buffer{}
	response.Write(&buf)
	resBytes := buf.Bytes()
	return resBytes, nil
}
