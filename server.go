package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"slices"
	"strconv"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
)

var (
	connections []*websocket.Conn
	mu          sync.Mutex
)

func pluralize(count int) string {
	if count == 1 {
		return ""
	}
	return "s"
}

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("🟨 could not read env file... using default config")
	}

	readBufferSizeStr := os.Getenv("READ_BUFFER_SIZE")
	writeBufferSizeStr := os.Getenv("WRITE_BUFFER_SIZE")
	readBufferSize := 1024
	writeBufferSize := 1024
	if readBufferSizeStr != "" {
		i, err := strconv.Atoi(readBufferSizeStr)
		if err == nil {
			readBufferSize = i
		}
	}
	if writeBufferSizeStr != "" {
		i, err := strconv.Atoi(writeBufferSizeStr)
		if err == nil {
			writeBufferSize = i
		}
	}

	var allowedOrigins string
	allowedOrigins = os.Getenv("ALLOWED_ORIGINS")

	var upgrader = websocket.Upgrader{
		ReadBufferSize:  readBufferSize,
		WriteBufferSize: writeBufferSize,
		CheckOrigin: func(r *http.Request) bool {
			if allowedOrigins == "" {
				// allow any origins
				return true
			} else {
				// allowed origins is a comma seperated string...
				// split into slice and user all defined
				origins := strings.Split(allowedOrigins, ",")
				if len(origins) == 1 {
					if r.Host == origins[0] {
						return true
					} else {
						return false
					}
				} else {
					if slices.Contains(origins, r.Host) {
						return true
					} else {
						return false
					}
				}
			}
		},
	}

	fmt.Printf("config\nallowed origins: %v\nreadBufferSize: %v\nwriteBufferSize: %v\n", allowedOrigins, readBufferSize, writeBufferSize)

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
