package server

import (
	"fmt"
	"errors"
	"sync"
	"log"
	"github.com/gorilla/websocket"
	
)

type Hub struct {
	mu         sync.Mutex
	conns      map[*websocket.Conn]string
	usernames  map[string]*websocket.Conn
}

func NewHub() *Hub {
	return &Hub{
		conns: make(map[*websocket.Conn]string),
		usernames: make(map[string]*websocket.Conn),
	}
}

func (h *Hub) add(conn *websocket.Conn,username string) error {
	h.mu.Lock()
	defer h.mu.Unlock()
  exists,ok := h.usernames[username]
	if ok && exists != nil {
		return errors.New("username already in use")
	}
	h.conns[conn] = username
	h.usernames[username] = conn
	return nil
}

func (h *Hub) remove(conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	username, ok := h.conns[conn]
	if ok {
		delete(h.usernames,username)
		delete(h.conns,conn)
	}
}

func (h *Hub) broadcast(messageType int, message []byte, username string) error {
	messageFinal := fmt.Sprintf("%s > %s" ,username,message)
	for conn := range h.conns {
		if username != h.conns[conn]{
			err := conn.WriteMessage(messageType, []byte(messageFinal))
			if err != nil {
				log.Printf("write error:%v", err)
			return err
			}
		}
	}
	return nil
}


func (h *Hub) connectionEcho(conn *websocket.Conn) {
	msg := "server > " + h.conns[conn] + " connected!"
	for connAux := range h.conns {
		if connAux != conn {
			connAux.WriteMessage(websocket.TextMessage,[]byte(msg))
		}
	}
}