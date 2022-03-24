package net

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"net/http"
)

type ResponseWriterWrapper struct {
	body       *bytes.Buffer
	statusCode *int
	header     http.Header
	conn       net.Conn
}

func NewResponseWriterWrapper(conn net.Conn, req *http.Request) ResponseWriterWrapper {
	var buf bytes.Buffer
	var statusCode int = 200
	return ResponseWriterWrapper{
		body:       &buf,
		statusCode: &statusCode,
		header:     req.Header,
		conn:       conn,
	}
}

// Hijack lets the caller take over the connection.
// After a call to Hijack the HTTP server library
// will not do anything else with the connection.
//
// It becomes the caller's responsibility to manage
// and close the connection.
//
// The returned net.Conn may have read or write deadlines
// already set, depending on the configuration of the
// Server. It is the caller's responsibility to set
// or clear those deadlines as needed.
//
// The returned bufio.Reader may contain unprocessed buffered
// data from the client.
//
// After a call to Hijack, the original Request.Body must not
// be used. The original Request's Context remains valid and
// is not canceled until the Request's ServeHTTP method
// returns.
func (rww ResponseWriterWrapper) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	clientBuf := bufio.NewReadWriter(bufio.NewReader(rww.conn), bufio.NewWriter(rww.conn))
	return rww.conn, clientBuf, nil
}

func (rww ResponseWriterWrapper) Write(buf []byte) (int, error) {
	return rww.body.Write(buf)
}

func (rww ResponseWriterWrapper) Header() http.Header {
	return rww.header

}

func (rww ResponseWriterWrapper) WriteHeader(statusCode int) {
	(*rww.statusCode) = statusCode
}

func (rww ResponseWriterWrapper) String() string {
	var buf bytes.Buffer

	buf.WriteString("Response:")

	buf.WriteString("Headers:")
	for k, v := range rww.header {
		buf.WriteString(fmt.Sprintf("%s: %v", k, v))
	}

	buf.WriteString(fmt.Sprintf(" Status Code: %d", *(rww.statusCode)))

	buf.WriteString("Body")
	buf.WriteString(rww.body.String())
	return buf.String()
}
