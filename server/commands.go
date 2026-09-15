package server

import (
	"fmt"
	"strings"
	"github.com/gorilla/websocket"
)

func handleAdminCommand(commandB []byte, conn *websocket.Conn) (response bool) {
	command := string(commandB)
	if !strings.HasPrefix(command,"!") {
		return false
	}
	conn.WriteMessage(websocket.TextMessage,[]byte("server > Admin password:"))
	_,adminPassByte,err := conn.ReadMessage()
	if err != nil {
		fmt.Printf("Unexpected error:",err)
		return true
	}
	if adminPass != string(adminPassByte) {
		conn.WriteMessage(websocket.TextMessage,[]byte("server > Incorrect password!"))
		return true
	}
	args := strings.Fields(command)
	if len (args) > 2 {
		conn.WriteMessage(websocket.TextMessage,[]byte("server > Too many arguments"))
		return true
	}
	switch args[0] {
	case "!kick":
		connAux := hub.usernames[args[1]]
		connAux.WriteMessage(websocket.TextMessage,[]byte("You have been kicked by an admin!"))
		disMsg := hub.conns[connAux] + " has been kicked!"
		hub.broadcast(websocket.TextMessage,[]byte(disMsg),"server")
		connAux.Close()
		return true
	default:
		return true
	}
}

func handleUserCommand(command []byte, conn *websocket.Conn) (response bool) {
	cmdStr := string(command)
	cmdStr = strings.ReplaceAll(cmdStr," ","")
	switch cmdStr {
	case "$exit":
		conn.Close()
		return true
	case "$list":
		for _,connL := range hub.conns {
			response := fmt.Sprintf("server > %s",connL)
			if connL != hub.conns[conn]{
				conn.WriteMessage(websocket.TextMessage,[]byte(response))
			}
		}
		return true
	default:
		return false
	}
	
}