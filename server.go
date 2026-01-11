package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"slices"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/urfave/cli/v3"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Dev  AppConfig `yaml:"dev"`
	Prod AppConfig `yaml:"prod"`
}

type AppConfig struct {
	AllowedOrigins  []string `yaml:"allowed_origins"`
	WriteBufferSize int      `yaml:"write_buffer_size"`
	ReadBufferSize  int      `yaml:"read_buffer_size"`
}

var (
	mu    sync.Mutex
	rooms = make(map[string][]*websocket.Conn) // KEY ROOM -> VALUE CONNECTIONS
)

func main() {
	cmd := &cli.Command{
		Name:        "gosocket",
		Description: "A lightweight Go-based CLI for interacting with WebSocket APIs",
		Flags: []cli.Flag{&cli.StringFlag{
			Name:     "mode",
			Usage:    "dev or prod",
			Required: true,
			Validator: func(s string) error {
				if s != "dev" && s != "prod" {
					return fmt.Errorf("not a valid mode")
				}
				return nil
			},
		}, &cli.StringFlag{
			Name:  "port",
			Value: "8080",
		}},
		Commands: []*cli.Command{
			{
				Name:  "start",
				Usage: "Starts the websocket server",
				Commands: []*cli.Command{
					{
						Name:  "dev",
						Usage: "Starts the websocket server in development mode",
						Action: func(ctx context.Context, cmd *cli.Command) error {
							mode := cmd.String("mode")
							port := cmd.String("port")

							config, err := initConfig(mode)
							if err != nil {
								return fmt.Errorf("%s", err.Error())
							}

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
								conn, err := upgrader.Upgrade(w, r, nil)
								if err != nil {
									log.Println(err)
									return
								}

								defer conn.Close()

								// join room
								roomName := r.URL.Path
								usersInRoom := rooms[roomName]
								usersInRoom = append(usersInRoom, conn)
								mu.Lock()
								rooms[roomName] = usersInRoom
								mu.Unlock()

								for {
									messageType, p, err := conn.ReadMessage()
									if err != nil {
										// leave room
										var updatedUsersInRoom []*websocket.Conn
										for _, v := range rooms[roomName] {
											if v == conn {
												continue
											}
											updatedUsersInRoom = append(updatedUsersInRoom, v)
										}
										mu.Lock()
										rooms[roomName] = updatedUsersInRoom
										mu.Unlock()

										log.Println(err)
										return
									}

									// send message back to all clients within this room
									for _, v := range rooms[roomName] {
										if err := v.WriteMessage(messageType, p); err != nil {
											log.Println(err)
											return
										}
									}
								}
							})

							log.Printf("websocket server listening on :%v\n", port)
							log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", port), nil))
							return nil
						},
					},
				},
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}

func initConfig(mode string) (AppConfig, error) {
	configFile := "config.yaml"

	// see if the config file exists
	// if not, create a default one for user
	_, err := os.Stat(configFile)
	if err != nil {
		var defaultConfig Config = Config{
			Dev: AppConfig{
				AllowedOrigins:  []string{},
				ReadBufferSize:  1024,
				WriteBufferSize: 1024,
			},
			Prod: AppConfig{
				AllowedOrigins:  []string{},
				ReadBufferSize:  1024,
				WriteBufferSize: 1024,
			},
		}

		yamlTxt, err := yaml.Marshal(&defaultConfig)
		if err != nil {
			return AppConfig{}, fmt.Errorf("there was an error while marshalling default yaml config")
		}
		os.WriteFile(configFile, yamlTxt, 0644)
		fmt.Println("initialzed defualt config yaml")
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		return AppConfig{}, fmt.Errorf("there was an error while reading config file")
	}

	var config Config
	if err = yaml.Unmarshal(data, &config); err != nil {
		return AppConfig{}, fmt.Errorf("error while reading config file")
	}

	// return the config for the mode passed in
	switch mode {
	case "dev":
		return config.Dev, nil
	case "prod":
		return config.Prod, nil
	default:
		return AppConfig{}, fmt.Errorf("%s", "cannot return config for mode "+mode)
	}
}
