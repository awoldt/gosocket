package socket

import (
	"slices"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

type SocketServer struct {
	Mu    sync.RWMutex
	Rooms map[string][]*websocket.Conn
}

func (s *SocketServer) Join(room string, socket *websocket.Conn) {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	// see if room already exists
	if _, ok := s.Rooms[room]; !ok {
		s.Rooms[room] = []*websocket.Conn{}
	}

	socketsInRoom := s.Rooms[room]
	socketsInRoom = append(socketsInRoom, socket)
	s.Rooms[room] = socketsInRoom
}

func (s *SocketServer) Leave(room string, socket *websocket.Conn) {
	for i, v := range s.Rooms[room] {
		if v == socket {
			s.Mu.Lock()
			defer s.Mu.Unlock()

			updatedRoom := slices.Delete(s.Rooms[room], i, i+1)

			// remove room from map if nobody in there anymore
			if len(updatedRoom) == 0 {
				delete(s.Rooms, room)
			} else {
				s.Rooms[room] = updatedRoom
			}
			break
		}
	}
}

func (s *SocketServer) SendMessage(room string, messageType int, message []byte, socket *websocket.Conn) {
	s.Mu.RLock()

	for _, conn := range s.Rooms[room] {
		if err := conn.WriteMessage(messageType, message); err != nil {
			logrus.Error(err)
			return
		}
	}

	s.Mu.RUnlock()
}
