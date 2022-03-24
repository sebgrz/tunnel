package helper

import (
	"io"
	"net"
)

func RecvBytes(conn net.Conn) ([]byte, error) {
	msgBytes := make([]byte, 0)
	for {
		b := make([]byte, 1024)
		bl, err := conn.Read(b)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		msgBytes = append(msgBytes, b[:bl]...)
		if bl < len(b) {
			break
		}
	}

	return msgBytes, nil
}
