package main

import (
	"io"
	"log"
	"net"
	"os"
)

func main() {
	l, err := net.Listen("tcp", ":5000")
	if err != nil {
		log.Fatal(err)
	}

	s := make(chan os.Signal)

	go func() {
		for {
			con, err := l.Accept()
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("New connection: %s", con.RemoteAddr().String())

			for {
				b := make([]byte, 1024)
				bl, err := con.Read(b)
				if err != nil {
					if err == io.EOF {
						break
					}
					log.Fatal(err)
				}
				log.Printf("Recv bytes: %d", bl)
				if bl == 0 {
					break
				}
				log.Println(string(b))
				if bl < len(b) {
					break
				}

			}
			log.Printf("End receiving")
			responseMessage := `
HTTP/1.1 200 OK
Date: Sun, 10 Oct 2010 23:26:07 GMT
Server: Proxy Server 
Last-Modified: Sun, 26 Sep 2010 22:04:35 GMT
ETag: "45b6-834-49130cc1182c0"
Accept-Ranges: bytes
Content-Length: 12
Connection: close
Content-Type: text/html

OK!
			`
			con.Write([]byte(responseMessage))
			con.Close()
		}
	}()

	<-s
	log.Print("server stopped")
}
