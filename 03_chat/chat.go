package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"

	"./logger"
)

const (
	telnetPortName = "telnetPort"
)

type Client struct {
	conn     io.ReadWriteCloser
	id string
	nickname string
	ch       chan chatMsg
}

type chatMsg struct {
	nick string
	ch   string
	msg  string
	ts   string
}

var log *logger.Log

func makeChatMessage(nick string, format string, args ...interface{}) chatMsg {

	return chatMsg{nick: nick, ts: time.Now().Format(time.Kitchen), msg: fmt.Sprintf(format, args...)}

}

//move this fmt string into config file
//
//found a page showing what this stuff means
//https://xdevs.com/guide/color_serial/
func formatChatMessage(c chatMsg) string {
	return fmt.Sprintf("\033[1;33;40m%s: %s \033[m:%s\033[m", c.ts, c.nick, c.msg)
}

func main() {

	var configDefault string = "chat"

	viper.SetConfigName(configDefault)
	viper.SetConfigType("toml")
	viper.AddConfigPath("./")
	viper.SetDefault(telnetPortName, "6000")

	err := viper.ReadInConfig()
	if err != nil {
		fmt.Printf("No configuration file found:%s:err:%v: - using defaults\n", configDefault, err)
	}

	logFile := viper.GetString("logFile")

	log = logger.SetupLoggingOrDie(logFile)

	log.Info("listening on port:%s:\n", viper.GetString(telnetPortName))
	ln, err := net.Listen("tcp", ":"+viper.GetString(telnetPortName))
	if err != nil {
		log.Error("Listener setup error:%v:\n", err)
		os.Exit(1)
	}

	msgchan := make(chan chatMsg)
	addchan := make(chan Client)
	rmchan := make(chan Client)

	go handleMessages(msgchan, addchan, rmchan)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Error("Listener accept error:%v:\n", err)
			continue
		}

		go handleConnection(conn, conn.RemoteAddr().String(), msgchan, addchan, rmchan)
	}

}

func (c Client) ReadLinesInto(ch chan chatMsg) {
	bufc := bufio.NewReader(c.conn)
	for {
		line, err := bufc.ReadString('\n')
		if err != nil {
			break
		}

		ch <- makeChatMessage(c.nickname, "%s", line)
	}
}

func (c Client) WriteLinesFrom(ch chan chatMsg) {
	for msg := range ch {
		_, err := io.WriteString(c.conn, formatChatMessage(msg))
		if err != nil {
			return
		}
	}
}

func promptNick(c io.ReadWriter, bufc *bufio.Reader) string {
	io.WriteString(c, "\033[1;30;41mWelcome to the fancy demo chat!\033[0m\n")
	io.WriteString(c, "What is your nick? ")
	nick, _, _ := bufc.ReadLine()
	return string(nick)
}

// telnet oriented
func handleConnection(c io.ReadWriteCloser, id string, msgchan chan chatMsg, addchan chan Client, rmchan chan Client) {

	bufc := bufio.NewReader(c)
	defer c.Close()
	client := Client{
		conn:     c,
		nickname: promptNick(c, bufc),
		ch:       make(chan chatMsg),
		id: id,
	}

	if strings.TrimSpace(client.nickname) == "" {
		io.WriteString(c, "Invalid Username\n")
		return
	}

	// Register user
	addchan <- client
	defer func() {
		msgchan <- makeChatMessage("system", "User %s left the chat room.\n", client.nickname)
		log.Info("Connection from %v closed.\n", id)
		rmchan <- client
	}()
	io.WriteString(c, fmt.Sprintf("Welcome, %s!\n\n", client.nickname))

	msgchan <- makeChatMessage("system", "New user %s has joined the chat room.\n", client.nickname)

	// I/O
	go client.ReadLinesInto(msgchan)
	client.WriteLinesFrom(client.ch)

}

func handleMessages(msgchan chan chatMsg, addchan chan Client, rmchan chan Client) {

	clients := make(map[string]chan chatMsg)

	for {
		select {
		case msg := <-msgchan:
			log.Info("New message: %s", msg.msg)
			for _, ch := range clients {
				go func(mch chan chatMsg, _msg chatMsg) { mch <- _msg }(ch, msg)
			}

		case client := <-addchan:
			// log.Info("New client: %v\n", client.conn.RemoteAddr())
			log.Info("New client:id:%v:\n", client.id)
			clients[client.id] = client.ch

		case client := <-rmchan:
			// log.Info("Client disconnects: %v\n", client.conn.RemoteAddr())
			log.Info("Client disconnects: %v\n", client.id)
			delete(clients, client.id)
		}
	}

}
