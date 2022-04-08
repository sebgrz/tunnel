package net

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
)

func ConfigureAndListen(port string, isSsl bool, config *tls.Config, listen func(conn net.Conn)) error {
	internalListener, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		return err
	}
	go func(listener net.Listener, listenCall func(conn net.Conn)) {
		for {
			con, err := listener.Accept()
			log.Printf("%v %v", con.LocalAddr(), con.RemoteAddr())
			if err != nil {
				log.Fatal(err)
			}

			if isSsl {
				conTls := tls.Server(con, config)
				listenCall(conTls)
			} else {
				listenCall(con)
			}
		}
	}(internalListener, listen)
	return nil
}
