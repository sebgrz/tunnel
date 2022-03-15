package http

import (
	"bufio"
	"bytes"
	"net/http"
)

func Send(destinationAddress string, msg []byte) ([]byte, error) {
	client := http.Client{}

	br := bufio.NewReader(bytes.NewReader(msg))
	request, err := http.ReadRequest(br)
	if err != nil {
		return nil, err
	}
	request.URL.Host = destinationAddress

	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	buf := bytes.Buffer{}
	response.Write(&buf)
	resBytes := buf.Bytes()
	return resBytes, nil
}
