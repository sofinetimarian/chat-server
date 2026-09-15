package server

import (
	"bufio"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"github.com/fatih/color"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var hub = NewHub()
var adminPass string = "admin"
var password string = ""
var port string = ":3333"
var upgrader websocket.Upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func logs() {
	logDir := "/logs" 
	if err := os.MkdirAll(logDir, 0755); err != nil {
		log.Panic(err)
	}

	logPath := filepath.Join(logDir, "logFile.log")

	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Panic(err)
	}
	defer f.Close()

	log.SetOutput(f)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("TEST LOG")
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("could not upgrade the connection", err)
		return
	}
	defer conn.Close()
	conn.WriteMessage(websocket.TextMessage, []byte("server > Enter the password:(leave empty if none)"))

	for i := 1; i <= 3; i++ {
		if i > 1 {
			conn.WriteMessage(websocket.TextMessage, []byte("server > Incorect password, Try again:"))
		}
		_, bytePass, err := conn.ReadMessage()
		if err != nil {
			log.Printf("unexpected error:%v", err)
			return
		}
		userPass := string(bytePass)
		if userPass == password {
			break
		}
		if i == 3 {
			conn.WriteMessage(websocket.TextMessage, []byte("server > Disconecting"))
			return
		}
	}
	conn.WriteMessage(websocket.TextMessage, []byte("server > Choose a username:"))
	_, byteUsername, err := conn.ReadMessage()
	if err != nil {
		log.Printf("unexpected error:%v", err)
		return
	}
	username := string(byteUsername)
	username = strings.ReplaceAll(username, " ", "")
	if username == "server" {
		conn.WriteMessage(websocket.TextMessage, []byte("server > username cannot be server"))
		return
	}
	err = hub.add(conn, username)
	if err != nil {
		conn.WriteMessage(websocket.TextMessage, []byte("server > "+err.Error()))
		return
	}
	defer hub.remove(conn)
	hub.connectionEcho(conn)
	conn.WriteMessage(websocket.TextMessage, []byte("server > Connected!"))
	log.Printf("Client Connected:", conn.RemoteAddr(), username)
	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err,
				websocket.CloseGoingAway,
				websocket.CloseAbnormalClosure) {
				log.Printf("Unexpected close error: %v", err)
			}
			break
		}
		log.Printf("Received: %s", message)
		res := handleUserCommand(message, conn)
		if res == true {
			continue
		}
		res = handleAdminCommand(message, conn)
		if res == true {
			continue
		}
		err = hub.broadcast(messageType, message, username)

		if err != nil {
			log.Printf("Unexpected close error:%v", err)
			break
		}
	}
	disMsg := hub.conns[conn] + " disconected!"
	hub.broadcast(websocket.TextMessage, []byte(disMsg), "server")
	log.Printf("Client disconected %s", conn.RemoteAddr())
}

func handleTest(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "test\n")
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

func checkPassword(command string) (bool, error) {
	if strings.Contains(command, " ") {
		return false, errors.New("password cannot contain spaces")
	}
	return true, nil
}

func NewServer() {
	reader := bufio.NewReader(os.Stdin)
	logs()
	var auxPort string
	var auxPassword string
	var auxAdminPass string
	color.Yellow("Please choose the server port (default:3333, press enter to use the default)")
	for {
		auxPort, _ = reader.ReadString('\n')
		auxPort = strings.TrimRight(auxPort, "\r\n")
		auxPort = strings.TrimLeft(auxPort, "\t")
		if auxPort == "exit" {
			os.Exit(0)
		}
		err, validPort := checkDigits(auxPort)
		if err != nil {
			color.Red("invalid port!")
		}
		if validPort {
			break
		}
	}
	if auxPort != "" {
		port = ":" + auxPort
	}
	color.Yellow("Please choose the server password (there is no password by default)")
	for {
		auxPassword, _ = reader.ReadString('\n')
		auxPassword = strings.TrimRight(auxPort, "\r\n")
		auxPassword = strings.TrimLeft(auxPort, "\t")
		validPass, err := checkPassword(auxPassword)
		if validPass {
			break
		}
		if err != nil {
			errMsg := "unexpeted error:" + err.Error()
			color.Red(errMsg)
		}
	}
	password = auxPassword
	color.Yellow("Please choose the admin password (default password:admin)")
	for {
		auxAdminPass, _ = reader.ReadString('\n')
		auxAdminPass = strings.TrimRight(auxPort, "\r\n")
		auxAdminPass = strings.TrimLeft(auxPort, "\t")
		if auxAdminPass == "" {
			break
		}
		validPass, err := checkPassword(auxAdminPass)
		if validPass {
			break
		}
		if err != nil {
			errMsg := "unexpeted error:" + err.Error()
			color.Red(errMsg)
		}
	}
	if auxAdminPass != "" {
		adminPass = auxAdminPass
	}
	w := mux.NewRouter()
	w.HandleFunc("/ws", handleWebSocket)
	w.HandleFunc("/test", handleTest)
	srv := &http.Server{
		Addr:    port,
		Handler: w,
	}
	go srv.ListenAndServe()
	succesMsg := "server started succesfully on port:" + port
	color.Green(succesMsg)
}
