package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/viper"

	m "./message"
	"./logger"
)

const (
	telnetPortName = "telnetPort"
	apiPortName = "apiPort"
)

type Client struct {
	conn     io.ReadWriteCloser
	id string
	nickname string
	ch       chan m.ChatMsg
	kind string
}

var log *logger.Log

//move this fmt string into config file
//
//found a page showing what this stuff means
//https://xdevs.com/guide/color_serial/
func formatChatMessage(c m.ChatMsg) string {
	return fmt.Sprintf("\033[1;33;40m%s: %s \033[m:%s\033[m", c.Time, c.Nick, c.Msg)
}

func main() {

	var configDefault string = "chat"

	viper.SetConfigName(configDefault)
	viper.SetConfigType("toml")
	viper.AddConfigPath("./")
	viper.SetDefault(telnetPortName, "6000")
	viper.SetDefault(apiPortName, "6001")

	err := viper.ReadInConfig()
	if err != nil {
		fmt.Printf("No configuration file found:%s:err:%v: - using defaults\n", configDefault, err)
	}

	logFile := viper.GetString("logFile")

	log = logger.SetupLoggingOrDie(logFile)

	log.Info("listening on ports:telnet:%s:api:%s:\n", viper.GetString(telnetPortName), viper.GetString(apiPortName))
	ln, err := net.Listen("tcp", ":" + viper.GetString(telnetPortName))
	if err != nil {
		log.Error("Listener setup error:%v:\n", err)
		os.Exit(1)
	}
	
	msgchan := make(chan m.ChatMsg)
	addchan := make(chan Client)
	rmchan := make(chan Client)

	go apiServer(viper.GetString(apiPortName), msgchan, addchan, rmchan)

	go handleMessages(msgchan, addchan, rmchan)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Error("Listener accept error:%v:\n", err)
			continue
		}

		go handleTelnetConnection(conn, conn.RemoteAddr().String(), msgchan, addchan, rmchan)
	}

}

func (c Client) ReadLinesInto(ch chan m.ChatMsg) {
	bufc := bufio.NewReader(c.conn)
	for {
		line, err := bufc.ReadString('\n')
		if err != nil {
			break
		}

		ch <- m.MakeChatMessage(c.nickname, "%s", line)
	}
}

func (c Client) WriteLinesFrom(ch chan m.ChatMsg) {
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
func handleTelnetConnection(c io.ReadWriteCloser, id string, msgchan chan m.ChatMsg, addchan chan Client, rmchan chan Client) {

	bufc := bufio.NewReader(c)
	defer c.Close()
	client := Client{
		conn:     c,
		nickname: promptNick(c, bufc),
		ch:       make(chan m.ChatMsg),
		id: id,
		kind: "telnet",
	}

	if strings.TrimSpace(client.nickname) == "" {
		io.WriteString(c, "Invalid Username\n")
		return
	}

	// Register user
	addchan <- client
	defer func() {
		msgchan <- m.MakeChatMessage("system", "User %s left the chat room.\n", client.nickname)
		log.Info("Connection from %v closed.\n", id)
		rmchan <- client
	}()
	io.WriteString(c, fmt.Sprintf("Welcome, %s!\n\n", client.nickname))

	msgchan <- m.MakeChatMessage("system", "New user %s has joined the chat room.\n", client.nickname)

	// I/O
	go client.ReadLinesInto(msgchan)
	client.WriteLinesFrom(client.ch)

}

func handleMessages(msgchan chan m.ChatMsg, addchan chan Client, rmchan chan Client) {

	clients := make(map[string]chan m.ChatMsg)

	for {
		select {
		case msg := <-msgchan:
			log.Info("New message: %s", msg.Msg)
			for _, ch := range clients {
				go func(mch chan m.ChatMsg, _msg m.ChatMsg) { mch <- _msg }(ch, msg)
			}

		case client := <-addchan:
			log.Info("New client:id:%v:\n", client.id)
			clients[client.id] = client.ch

		case client := <-rmchan:
			log.Info("Client disconnects: %v\n", client.id)
			delete(clients, client.id)
		}
	}

}

func apiServer(port string, msgchan chan m.ChatMsg, addchan chan Client, rmchan chan Client) {

	http.HandleFunc("/chat/", func(w http.ResponseWriter, req *http.Request) {

		var channel, nick, msg string

		urlParts := strings.Split(req.URL.Path, "/")

		log.Info("api call:%s:parts:%d:\n", req.URL.Path, len(urlParts))

		if len(urlParts) < 5 {
	        http.NotFound(w, req)
	        return
        } else if urlParts[3] == "" {
	        http.NotFound(w, req)
	        return
        } else if urlParts[4] == "" {
	        http.NotFound(w, req)
	        return
		} else {
			channel = urlParts[2]
			nick = urlParts[3]
			msg = strings.Join(urlParts[4:], "/")
		}

		log.Info("api call:channel:%s:nick:%s:msg:%s:\n", channel, nick, msg)

		msgchan <- m.MakeChatMessage(nick, "%s\n", msg)

		fmt.Fprintf(w, "sending message for:user:%s:to chan:%s:\n", nick, channel)

	})

	err := http.ListenAndServe(":" + port, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}

}