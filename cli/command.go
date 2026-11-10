package cli

import (
	"context"
	"fmt"
	"net/http"

	"github.com/awoldt/gosocket/socket"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v3"
)

var StartCommand = &cli.Command{
	Description: "A lightweight websocket server written in Go",
	Flags: []cli.Flag{
		PortFlag,
		LoggingFlag,
	},
	Action: func(ctx context.Context, c *cli.Command) error {
		var rooms socket.SocketServer = socket.SocketServer{
			Rooms: make(map[string][]*websocket.Conn),
		}

		port := c.String("port")
		logrus.Infof("websocket server listening on :%v", port)

		var upgrader = websocket.Upgrader{}

		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			conn, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				logrus.Error(err)
				return
			}

			defer conn.Close()

			// join room
			roomName := r.URL.Path
			rooms.Join(roomName, conn)
			logrus.Info("someone has joined room " + roomName)

			for {
				messageType, p, err := conn.ReadMessage()
				if err != nil {
					// leave room
					rooms.Leave(roomName, conn)
					logrus.Info("someone has left room " + roomName)
					return
				}

				// send message
				rooms.SendMessage(roomName, messageType, p, conn)
			}
		})

		logrus.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", port), nil))
		return nil
	},
}
