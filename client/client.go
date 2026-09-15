package client

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/gorilla/websocket"
)

var url string

func readLoop(conn *websocket.Conn, done chan struct{}) {
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Printf("read error:%v", err)
			close(done)
			return
		}
		color.Green(string(msg))
	}
}

func writeLoop(conn *websocket.Conn, reader *bufio.Reader, done chan struct{}) {
	for {
		msg, err := reader.ReadString('\n')
		if err != nil {
			close(done)
			fmt.Printf("unexpected error:%v", err)
			return
		}
		msg = strings.TrimRight(msg, "\r\n")
		if msg == "$exit" {
			close(done)
			return
		}
		err = conn.WriteMessage(websocket.TextMessage, []byte(msg))
		if err != nil {
			close(done)
			fmt.Printf("unexpected error:%v", err)
			return
		}
	}
}

func checkDigits(command string) (error, bool) {
	if command == "" {
		return nil, true
	}
	if strings.Contains(command, " ") {
		return errors.New("invalid port"), false
	}
	for _, char := range command {
		if !strings.ContainsRune("0123456789", char) {
			return errors.New("invalid port"), false
		}
	}
	return nil, true
}

func NewClient() {
	reader := bufio.NewReader(os.Stdin)
	color.Yellow("IP address:")
	var ip string
	var port string
	ip, _ = reader.ReadString('\n')
	ip = strings.TrimRight(ip, "\r\n")
	color.Yellow("Port:")
	port, _ = reader.ReadString('\n')
	port = strings.TrimRight(port, "\r\n")
	url = "ws://" + ip + ":" + port + "/ws"
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Fatal("dial:", err)
		return
	}
	defer conn.Close()

	done := make(chan struct{})

	go readLoop(conn, done)
	writeLoop(conn, reader, done)

	<-done
}
