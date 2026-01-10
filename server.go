package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/urfave/cli/v3"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Apps map[string]App `yaml:"apps"`
}
type App struct {
	Url  string    `yaml:"url"`
	Dev  AppConfig `yaml:"dev"`
	Prod AppConfig `yaml:"prod"`
}
type AppConfig struct {
	AllowedOrigins  string `yaml:"allowed_origins"`
	WriteBufferSize int    `yaml:"write_buffer_size"`
	ReadBufferSize  int    `yaml:"read_buffer_size"`
}

var (
	connections []*websocket.Conn
	mu          sync.Mutex
)

func main() {
	_, err := initConfig()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	cmd := &cli.Command{
		Name: "websoget",
		Flags: []cli.Flag{&cli.StringFlag{
			Name:     "mode",
			Usage:    "dev or prod",
			Required: true,
		}, &cli.StringFlag{
			Name:  "port",
			Value: "8080",
		}},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			mode := cmd.String("mode")
			if mode != "dev" && mode != "prod" {
				return fmt.Errorf("not a valid mode")
			}

			fmt.Printf("running websocket server in %v mode", mode)

			initConfig()

			var upgrader = websocket.Upgrader{
				ReadBufferSize:  1024,
				WriteBufferSize: 1024,
				CheckOrigin: func(r *http.Request) bool {
					return true
				},
			}

			http.HandleFunc("/socket", func(w http.ResponseWriter, r *http.Request) {
				conn, err := upgrader.Upgrade(w, r, nil)
				if err != nil {
					log.Println(err)
					return
				}

				defer conn.Close()

				// add connection
				mu.Lock()
				connections = append(connections, conn)
				mu.Unlock()

				// notify everyone about new user
				mu.Lock()
				userCount := len(connections)
				for _, c := range connections {
					c.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("New user joined! %d user%s online", userCount, "RAWR")))
				}
				mu.Unlock()

				for {
					messageType, message, err := conn.ReadMessage()
					if err != nil {
						log.Println(err)
						// remove connection on error
						mu.Lock()
						for i, c := range connections {
							if c == conn {
								connections = append(connections[:i], connections[i+1:]...)
								break
							}
						}
						userCount := len(connections)
						// notify remaining users
						for _, c := range connections {
							c.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("User left. %d user%s online", userCount, "rawr")))
						}
						mu.Unlock()
						return
					}

					// broadcast message to everyone
					mu.Lock()
					for _, c := range connections {
						if err := c.WriteMessage(messageType, message); err != nil {
							log.Println(err)
						}
					}
					mu.Unlock()
				}

			})

			log.Println("websocket server listening on :8080")
			log.Fatal(http.ListenAndServe(":8080", nil))
			return nil
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}

func initConfig() (Config, error) {
	data, err := os.ReadFile("config.yaml")
	if err != nil {
		return Config{}, fmt.Errorf("could not find config file")
	}

	var config Config

	if err = yaml.Unmarshal(data, &config); err != nil {
		return Config{}, fmt.Errorf("error while reading config file")
	}

	return config, nil
}
