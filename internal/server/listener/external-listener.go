package listener

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"log"
	"net"
	"net/http"
	"proxy/internal/server/configuration"
	"proxy/internal/server/connection"
	"proxy/internal/server/enum"
	"proxy/internal/server/inter"
	"proxy/internal/server/pack"
	"proxy/pkg/helper"
	"proxy/pkg/key"
	pkgnet "proxy/pkg/net"
	"strings"
	"sync"
)

type ExternalListener struct {
	config               *configuration.Configuration
	port                 string
	sslPort              string
	connections          map[string]inter.ExternalConnection
	chanAddConnection    chan pack.ChanExternalConnection
	chanRemoveConnection chan string
	chanMsgToInternal    chan<- pack.ChanProxyMessageToInternal
	chanMsgToExternal    <-chan pack.ChanProxyMessageToExternal
}

func NewExternalListener(
	config *configuration.Configuration,
	port string,
	sslPort string,
	chanMsgToInternal chan<- pack.ChanProxyMessageToInternal,
	chanMsgToExternal <-chan pack.ChanProxyMessageToExternal,
	chanAgentConnectionClosedToExternal <-chan string,
	chanExternalConnectionClosedToAgent chan<- string,
) *ExternalListener {
	l := &ExternalListener{
		config:               config,
		port:                 port,
		sslPort:              sslPort,
		connections:          make(map[string]inter.ExternalConnection),
		chanAddConnection:    make(chan pack.ChanExternalConnection),
		chanRemoveConnection: make(chan string),
		chanMsgToInternal:    chanMsgToInternal,
		chanMsgToExternal:    chanMsgToExternal,
	}

	go func() {
		mu := sync.Mutex{}
		for {
			select {
			case msgToExternal := <-chanMsgToExternal:
				mu.Lock()
				if con, ok := l.connections[msgToExternal.ExternalConnectionID]; ok {
					err := con.Send(msgToExternal.ExternalConnectionID, msgToExternal.Content)
					if err != nil {
						log.Print(err)
					}
				}
				mu.Unlock()
			case addConnection := <-l.chanAddConnection:
				l.connections[addConnection.ConnectionID] = addConnection.Connection
				log.Printf("external connection: %s added", addConnection.ConnectionID)
			case removeConnection := <-l.chanRemoveConnection:
				mu.Lock()
				if con, ok := l.connections[removeConnection]; ok {
					go func(externalConnection inter.ExternalConnection) {
						l.chanMsgToInternal <- pack.ChanProxyMessageToInternal{
							ExternalConnectionID: externalConnection.GetID(),
							Host:                 externalConnection.GetHost(),
							ConnectionType:       externalConnection.GetConnectionType(),
							MessageType:          enum.CloseConnectionExternalToInternalMessageType,
							Content:              []byte(""),
						}
					}(con)
				}
				delete(l.connections, removeConnection)
				mu.Unlock()
				// TODO: send message about close connection
				// In order to inform agent which connection has been closed and then close agent<->destination
				// connection
				// TODO: remember bottom case - in this case close connection information shouldn't be send
				log.Printf("external connection: %s removed", removeConnection)
			case connectionID := <-chanAgentConnectionClosedToExternal:
				if conn, ok := l.connections[connectionID]; ok {
					conn.Close()
				}
				go func() { l.chanRemoveConnection <- connectionID }()
			}
			log.Printf("external connections: %d", len(l.connections))
		}
	}()
	return l
}

func (l *ExternalListener) Run() {
	listen := func(con net.Conn) {
		connType, firstRecvBytes, recognizedRequest, err := recognizeConnectionType(con)

		if err != nil {
			log.Println(err)
		}

		log.Println(connType)

		var externalConnection inter.ExternalConnection
		switch connType {
		case connection.HTTPExternalConnectionType:
			externalConnection = connection.NewHTTPExternalConnection(con, l.chanRemoveConnection, l.chanMsgToInternal)
			l.chanAddConnection <- pack.ChanExternalConnection{
				ConnectionID: externalConnection.GetID(),
				Connection:   externalConnection,
			}
		case connection.WSExternalConnectionType:
			externalConnection = connection.NewWSExternalConnection(con, l.chanRemoveConnection, l.chanMsgToInternal)
			l.chanAddConnection <- pack.ChanExternalConnection{
				ConnectionID: externalConnection.GetID(),
				Connection:   externalConnection,
			}
		}
		externalConnection.InitialData(&inter.ExternalConnectionInitialData{
			MsgBytes: firstRecvBytes,
			Request:  recognizedRequest,
		})
		go externalConnection.Listen()
	}

	// Unencrypted traffic
	if l.port != "" {
		if err := pkgnet.ConfigureAndListen(l.port, false, nil, listen); err != nil {
			log.Fatal(err)
		}
	}
	// Encrypted traffic
	if l.sslPort != "" {
		if l.config == nil || l.config.Certificates == nil {
			log.Fatal("configuration is empty")
		}

		certificates := []tls.Certificate{}
		for _, certConfig := range l.config.Certificates {
			cer, err := tls.LoadX509KeyPair(certConfig.CertPath, certConfig.CertKeyPath)
			if err != nil {
				log.Fatal(err)
			}
			certificates = append(certificates, cer)
		}

		sslConfig := &tls.Config{Certificates: certificates}
		if err := pkgnet.ConfigureAndListen(l.sslPort, true, sslConfig, listen); err != nil {
			log.Fatal(err)
		}
	}
}

func recognizeConnectionType(conn net.Conn) (connection.ExternalConnectionType, *[]byte, *http.Request, error) {
	msgBytes, err := helper.RecvBytes(conn)
	if err != nil {
		return -1, nil, nil, err
	}
	br := bufio.NewReader(bytes.NewReader(msgBytes))
	request, err := http.ReadRequest(br)
	if err != nil {
		return -1, nil, nil, err
	}
	connectionHeader := request.Header.Get(key.ConnectionHTTPHeader)
	upgradeHeader := request.Header.Get(key.UpgradeHTTPHeader)

	if strings.ToLower(connectionHeader) == "upgrade" && strings.ToLower(upgradeHeader) == "websocket" {
		return connection.WSExternalConnectionType, &msgBytes, request, nil
	}
	return connection.HTTPExternalConnectionType, &msgBytes, request, nil
}
