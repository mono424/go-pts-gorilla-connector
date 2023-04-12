package ptsc_gorilla

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/mono424/go-pts"
)

func NewConnector(upgrader websocket.Upgrader, errorHandler pts.ErrorHandlerFunc) *pts.Connector {
	var connector *pts.Connector
	connector = pts.NewConnector(
		func(writer http.ResponseWriter, request *http.Request, properties map[string]interface{}) error {
			mutex := sync.Mutex{}

			conn, err := upgrader.Upgrade(writer, request, nil)
			if err != nil {
				return err
			}

			client := connector.Join(
				func(message []byte) error {
					mutex.Lock()
					defer mutex.Unlock()
					return conn.WriteMessage(websocket.TextMessage, message)
				},
				properties,
			)

			go func() {
				defer (func() {
					err := conn.Close()
					if err != nil {
						fmt.Println("Error during closing connection:", err)
					}
					connector.Leave(client.Id)
				})()

				for {
					_, message, err := conn.ReadMessage()
					if err != nil {
						switch err.(type) {
						case *websocket.CloseError:
							// omit ual websocket close errors
						default:
							fmt.Println("Error during message reading:", err)
						}
						break
					}
					connector.Message(client.Id, message)
				}
			}()

			return nil
		},
		errorHandler,
	)

	return connector
}
