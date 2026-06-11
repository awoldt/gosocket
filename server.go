package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"slices"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v3"
)

type Config struct {
	AllowedOrigins  []string
	WriteBufferSize int
	ReadBufferSize  int
	AuthToken       string
	Port            string
	Logging         bool
}

var (
	mu    sync.RWMutex
	rooms = make(map[string][]*websocket.Conn) // KEY ROOM -> VALUE CONNECTIONS
)

func main() {
	cmd := &cli.Command{
		Name:        "gosocket",
		Description: "A lightweight Go-based CLI for interacting with WebSocket APIs",
		Commands: []*cli.Command{
			{
				Name:  "start",
				Usage: "Starts the websocket server",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "port",
						Usage: "port for the websocket server to listen on",
					},
					&cli.StringSliceFlag{
						Name:  "allowed_origins",
						Usage: "origins allowed to connect",
					},
					&cli.IntFlag{
						Name:  "read_buffer_size",
						Usage: "websocket read buffer size in bytes",
					},
					&cli.IntFlag{
						Name:  "write_buffer_size",
						Usage: "websocket write buffer size in bytes",
					},
					&cli.StringFlag{
						Name:  "auth_token",
						Usage: "token required for client connections",
					},
					&cli.BoolFlag{
						Name:  "logs",
						Usage: "Enables server-side logging to the console",
					},
				},
				Action: func(ctx context.Context, c *cli.Command) error {
					config := Config{}
					config = applyConfigFlags(config, c)
					port := config.Port

					// set logging
					if !config.Logging {
						logrus.SetFormatter(&logrus.TextFormatter{
							FullTimestamp:   true,
							TimestampFormat: "2006-01-02 15:04:05",
						})
					} else {
						fmt.Printf("websocket server listening on port :%v", config.Port)
						logrus.SetOutput(io.Discard)
					}

					logrus.Infof("websocket server listening on :%v", port)

					var upgrader = websocket.Upgrader{
						ReadBufferSize:  config.ReadBufferSize,
						WriteBufferSize: config.WriteBufferSize,
						CheckOrigin: func(r *http.Request) bool {
							// if no origins set, allow all
							if len(config.AllowedOrigins) == 0 {
								return true
							}
							if slices.Contains(config.AllowedOrigins, r.Host) {
								return true
							}
							return false
						},
					}

					http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
						// auth
						if config.AuthToken != "" {
							token := r.URL.Query().Get("token")
							if token == "" || token != config.AuthToken {
								w.WriteHeader(401)
								w.Write([]byte("Unauthorized"))
								return
							}
						}

						conn, err := upgrader.Upgrade(w, r, nil)
						if err != nil {
							logrus.Error(err)
							return
						}

						defer conn.Close()

						// join room
						roomName := r.URL.Path
						mu.Lock()
						usersInRoom := rooms[roomName]
						usersInRoom = append(usersInRoom, conn)
						rooms[roomName] = usersInRoom
						mu.Unlock()

						logrus.Info("someone has joined room " + roomName)

						for {
							messageType, p, err := conn.ReadMessage()
							if err != nil {
								// leave room
								updatedUsersInRoom := []*websocket.Conn{}

								mu.Lock()
								for _, v := range rooms[roomName] {
									if v == conn {
										continue
									}
									updatedUsersInRoom = append(updatedUsersInRoom, v)
								}

								if len(updatedUsersInRoom) == 0 {
									// nobody in room anymore... just delete room from map
									delete(rooms, roomName)
								} else {
									rooms[roomName] = updatedUsersInRoom
								}
								mu.Unlock()

								logrus.Info("someone has left room " + roomName)
								return
							}

							// send message back to all clients within this room
							// DONT lock while WriteMessage is going (can cause bad performance for slow clients)
							// lock, create a copy of connections slice, unlock, THEN WriteMessage
							mu.RLock()
							conns := append(make([]*websocket.Conn, 0), rooms[roomName]...)
							mu.RUnlock()
							for _, v := range conns {
								if err := v.WriteMessage(messageType, p); err != nil {
									logrus.Error(err)
									return
								}
							}

						}
					})

					logrus.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", port), nil))
					return nil
				},
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		logrus.Fatal(err)
	}
}

func applyConfigFlags(config Config, c *cli.Command) Config {
	if c.IsSet("port") {
		config.Port = c.String("port")
	} else {
		config.Port = "8080"
	}

	if c.IsSet("allowed_origins") {
		config.AllowedOrigins = c.StringSlice("allowed_origins")
	}

	if c.IsSet("read_buffer_size") {
		config.ReadBufferSize = c.Int("read_buffer_size")
	} else {
		config.ReadBufferSize = 1024
	}

	if c.IsSet("write_buffer_size") {
		config.WriteBufferSize = c.Int("write_buffer_size")
	} else {
		config.WriteBufferSize = 1024
	}

	if c.IsSet("auth_token") {
		config.AuthToken = c.String("auth_token")
	}

	if c.IsSet("logs") {
		config.Logging = c.Bool("logs")
	} else {
		config.Logging = false
	}

	return config
}
