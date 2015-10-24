package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"time"

	"github.com/op/go-logging"
	"github.com/spf13/viper"
)

const (
	telnetPortName = "telnetPort"
)

type Client struct {
	conn     net.Conn
	nickname string
	ch       chan chatMsg
}

type chatMsg struct {
	ch string
	msg string
	ts string
}

// logger used throughout
// defaults to stdout
// setup in setupLoggingOrDie
var log *logging.Logger

// setup logging properly or die
// logs are not open yet so write for Std*
func setupLoggingOrDie(logFile string) *logging.Logger {

	//default log to stdout
	var logHandle io.WriteCloser = os.Stdout

	var err error

	if logFile != "" {

		if logHandle, err = os.OpenFile(logFile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666); err != nil {
			fmt.Fprintf(os.Stderr, "Can't open log:%s:err:%v:\n", logFile, err)
			os.Exit(1)
		}

		fmt.Printf("Logging to:logFile:%s:\n", logFile)

	} else {
		fmt.Printf("No logfile specified - going with stdout\n")
	}

	_log, err := logging.GetLogger("chatLog")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Can't start logger:%s:err:%v:\n", logFile, err)
		os.Exit(1)
	}

	backend1 := logging.NewLogBackend(logHandle, "", 0)
	backend1Leveled := logging.AddModuleLevel(backend1)
	backend1Leveled.SetLevel(logging.INFO, "")
	logging.SetBackend(backend1Leveled)

	return _log

}

func makeChatMessage(format string, args ...interface{}) chatMsg {

	return chatMsg{ ts: time.Now().Format(time.Kitchen), msg: fmt.Sprintf(format, args...) }

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

	log = setupLoggingOrDie(logFile)

	log.Info("listening on port:%s:\n", viper.GetString(telnetPortName))
	ln, err := net.Listen("tcp", ":"+ viper.GetString(telnetPortName))
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

		go handleConnection(conn, msgchan, addchan, rmchan)
	}
}

func (c Client) ReadLinesInto(ch chan<- chatMsg) {
	bufc := bufio.NewReader(c.conn)
	for {
		line, err := bufc.ReadString('\n')
		if err != nil {
			break
		}

		ch <- makeChatMessage("%s: %s", c.nickname, line)
	}
}

func (c Client) WriteLinesFrom(ch <-chan chatMsg) {
	for msg := range ch {
		_, err := io.WriteString(c.conn, msg.msg)
		if err != nil {
			return
		}
	}
}

func promptNick(c net.Conn, bufc *bufio.Reader) string {
	io.WriteString(c, "\033[1;30;41mWelcome to the fancy demo chat!\033[0m\n")
	io.WriteString(c, "What is your nick? ")
	nick, _, _ := bufc.ReadLine()
	return string(nick)
}

func handleConnection(c net.Conn, msgchan chan<- chatMsg, addchan chan<- Client, rmchan chan<- Client) {

	bufc := bufio.NewReader(c)
	defer c.Close()
	client := Client{
		conn:     c,
		nickname: promptNick(c, bufc),
		ch:       make(chan chatMsg),
	}

	if strings.TrimSpace(client.nickname) == "" {
		io.WriteString(c, "Invalid Username\n")
		return
	}

	// Register user
	addchan <- client
	defer func() {
		msgchan <- makeChatMessage("User %s left the chat room.\n", client.nickname)
		log.Info("Connection from %v closed.\n", c.RemoteAddr())
		rmchan <- client
	}()
	io.WriteString(c, fmt.Sprintf("Welcome, %s!\n\n", client.nickname))

	msgchan <- makeChatMessage("New user %s has joined the chat room.\n", client.nickname)

	// I/O
	go client.ReadLinesInto(msgchan)
	client.WriteLinesFrom(client.ch)

}

func handleMessages(msgchan <-chan chatMsg, addchan <-chan Client, rmchan <-chan Client) {

	clients := make(map[net.Conn]chan<- chatMsg)

	for {
		select {
		case msg := <-msgchan:
			log.Info("New message: %s", msg.msg)
			for _, ch := range clients {
				// go func(mch chan<- chatMsg) { mch <- "\033[1;33;40m" + msg.ts + ":" + msg.msg + "\033[m" }(ch)
				go func(mch chan<- chatMsg) { mch <- makeChatMessage("\033[1;33;40m" + msg.ts + ":" + msg.msg + "\033[m") }(ch)
			}

		case client := <-addchan:
			log.Info("New client: %v\n", client.conn.RemoteAddr())
			clients[client.conn] = client.ch

		case client := <-rmchan:
			log.Info("Client disconnects: %v\n", client.conn.RemoteAddr())
			delete(clients, client.conn)
		}
	}

}
