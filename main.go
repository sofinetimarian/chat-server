package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"chat-server/server"
	"chat-server/client"	
	"github.com/common-nighthawk/go-figure"
	"github.com/fatih/color"
)

func clearScreen() {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "cls")
	} else {
		cmd = exec.Command("clear")
	}
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func main() {
	clearScreen()
	reader := bufio.NewReader(os.Stdin)
	logo := figure.NewFigure("Simple Chat Server","",true)
	logo.Print()
	color.Yellow("Enter the \"help\" command for information on how to start or connect to the server")
	for {
		fmt.Print("> ")
		command,err := reader.ReadString('\n')
		command = strings.ReplaceAll(command," ","")
		if err != nil {
			fmt.Print("unexpected error:",err)
		}
		command = strings.Trim(command,"\r\n")
		if command == "exit" {
			clearScreen()
			os.Exit(0)
		} else if command == "clear" {
			clearScreen()
		} else if command == "help" {
			color.Cyan("connect - this command is used the connect to a server as the name implies, follow the steps provided after the command execution to succesfully connect to a server")
			color.Cyan("start - this command is used to start a chat server, follow the instructions provided after execution")
			color.Cyan("clear - this command is used to clear the screen")
			color.Cyan("exit - this command is used to exit the program")
		} else if command == "start" {
			server.NewServer()
		} else if command == "connect" {
			clearScreen()
			client.NewClient()
		}else {
			msg := command + ":unknown command"
			color.Red(msg)
		}
	}
}