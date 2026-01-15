package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"slices"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v3"
	"gopkg.in/yaml.v3"
)

type Config struct {
	AllowedOrigins  []string `yaml:"allowed_origins"`
	WriteBufferSize int      `yaml:"write_buffer_size"`
	ReadBufferSize  int      `yaml:"read_buffer_size"`
	AuthToken       string   `yaml:"auth_token"`
	Port            string   `yaml:"port"`
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
				Action: func(ctx context.Context, c *cli.Command) error {
					config, err := initConfig()
					if err != nil {
						logrus.Error("there was an error while initializing config")
						return fmt.Errorf("%s", err.Error())
					}

					logrus.SetFormatter(&logrus.TextFormatter{
						FullTimestamp:   true,
						TimestampFormat: "2006-01-02 15:04:05",
					})

					port := config.Port

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

					http.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
						// count the number of rooms and all connections
						htmlPage := buildStatsPage()

						w.Write([]byte(htmlPage))
					})

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

					logrus.Infof("websocket server listening on :%v\n", port)
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

func initConfig() (Config, error) {
	configFile := "config.yaml"

	// see if the config file exists
	// if not, create a default one for user
	_, err := os.Stat(configFile)
	if err != nil {
		var defaultConfig Config = Config{
			AllowedOrigins:  []string{},
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			Port:            "8080",
			AuthToken:       "",
		}

		yamlTxt, err := yaml.Marshal(&defaultConfig)
		if err != nil {
			return Config{}, fmt.Errorf("there was an error while marshalling default yaml config")
		}
		os.WriteFile(configFile, yamlTxt, 0644)
		logrus.Info("initialized default config yaml")
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		return Config{}, fmt.Errorf("there was an error while reading config file")
	}

	var config Config
	if err = yaml.Unmarshal(data, &config); err != nil {
		return Config{}, fmt.Errorf("error while reading config file")
	}

	return config, nil
}
