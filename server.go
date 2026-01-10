package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		if r.Host == "localhost:8080" {
			return true
		} else {
			return false
		}
	},
}

var (
	connections []*websocket.Conn
	mu          sync.Mutex
)

func main() {
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

		for {
			messageType, _, err := conn.ReadMessage()
			if err != nil {
				log.Println(err)
				return
			}

			// send to everyone
			for _, c := range connections {
				if err := c.WriteMessage(messageType, []byte(fmt.Sprintf("There are %v users", len(connections)))); err != nil {
					log.Println(err)
					return
				}
			}
		}
	})

	log.Println("websocket server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
