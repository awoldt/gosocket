package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/gorilla/websocket"
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
			c.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("New user joined! %d user%s online", userCount, pluralize(userCount))))
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
					c.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("User left. %d user%s online", userCount, pluralize(userCount))))
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
}

func pluralize(count int) string {
	if count == 1 {
		return ""
	}
	return "s"
}

func initConfig() {
	data, err := os.ReadFile("config.yaml")
	if err != nil {
		fmt.Println("no config found")
	}

	fmt.Println(string(data))

	var config Config

	if err = yaml.Unmarshal(data, &config); err != nil {
		fmt.Println(err.Error())
		panic("error while reading yaml config")
	}

	fmt.Println(config.Apps)
}
