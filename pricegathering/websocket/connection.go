package websocket

import (
	"encoding/json"
	"log"

	"github.com/gorilla/websocket"
)

type WebSocketConnection interface {
	connect() error
	subscribeToChannels(channels []interface{}) error
	readMessages() <-chan []byte
	disconnect() error
}

type WebSocket struct {
	SubscribeMessage interface{}
	Data             <-chan []byte
}

func NewWebSocket(websocketURL string, subscribeMessage interface{}) *WebSocket {
	w := &WebSocket{SubscribeMessage: subscribeMessage}
	con, err := w.connect(websocketURL)
	if err != nil {
		log.Println("asd")
	}
	_ = w.subscribeToChannels(con)
	w.Data = w.readMessages(websocketURL, con)
	return w
}

func (f *WebSocket) connect(websocketURL string) (*websocket.Conn, error) {
	conn, _, err := websocket.DefaultDialer.Dial(websocketURL, nil)
	if err != nil {
		log.Printf("Error establishing WebSocket connection: %v", err)
		return nil, err
	}
	return conn, nil
}

func (f *WebSocket) subscribeToChannels(conn *websocket.Conn) error {
	data, err := json.Marshal(f.SubscribeMessage)
	if err != nil {
		log.Printf("Error marshaling subscription message: %v", err)
		return err
	}

	err = conn.WriteMessage(websocket.TextMessage, data)
	if err != nil {
		log.Printf("Error sending subscription message: %v", err)
		return err
	}
	return nil
}

func (f *WebSocket) readMessages(websocketURL string, conn *websocket.Conn) <-chan []byte {
	messages := make(chan []byte)

	go func() {
		defer close(messages)
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Printf("Error reading WebSocket message: %v", err)

				newConn, err := f.connect(websocketURL)
				if err != nil {
					log.Printf("Error reconnecting to WebSocket: %v", err)
					continue
				}

				conn.Close()
				conn = newConn
			}
			if message != nil {
				messages <- message
			}
		}
	}()

	return messages
}

func (f *WebSocket) disconnect(conn *websocket.Conn) error {
	err := conn.Close()
	if err != nil {
		log.Printf("Error closing WebSocket connection: %v", err)
	}
	return err
}
