package ptsc

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/mono424/go-pts"
)

func NewConnector(upgrader websocket.Upgrader, errorHandler pts.ErrorHandlerFunc) *pts.Connector {
	var connector *pts.Connector
	connector = pts.NewConnector(
		func(writer http.ResponseWriter, request *http.Request, properties map[string]interface{}) error {
			conn, err := upgrader.Upgrade(writer, request, nil)
			if err != nil {
				return err
			}

			client := connector.Join(
				func(message []byte) error {
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
						fmt.Println("Error during message reading:", err)
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
