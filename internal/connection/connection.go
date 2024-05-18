package connection

import (
	"fmt"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

type Connection interface {
	CreateWebsocketConnection(url string) *websocket.Conn
}

type WebSocketConnection struct{}

func NewWebsocketConnection() *WebSocketConnection {
	return &WebSocketConnection{}
}

func (c *WebSocketConnection) CreateWebsocketConnection(url string) *websocket.Conn {
	var connection *websocket.Conn
	var err error
	for i := 0; i < 3; i++ {
		retryDuration := (i + 1) * 1
		connection, _, err = websocket.DefaultDialer.Dial(url, nil)
		if err == nil {
			fmt.Println("connected to websocket")
			return connection
		}
		log.Printf("unable to connect to websocket, retrying in %d seconds: %v", retryDuration, err)
		time.Sleep(time.Duration(retryDuration) * time.Second)
	}

	panic(err)
}
