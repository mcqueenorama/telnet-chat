package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"

	// "github.com/op/go-logging"
	"github.com/spf13/viper"
)

type Client struct {
	conn     net.Conn
	nickname string
	ch       chan string
}

func main() {

	var configDefault string = "chat"

	viper.SetConfigName(configDefault)
	viper.SetConfigType("toml")
	viper.AddConfigPath("./")
	viper.SetDefault("port", "6000")
	
	err := viper.ReadInConfig()
	if err != nil {
		log.Printf("No configuration file found:%s:err:%v: - using defaults\n", configDefault, err)
	}


	log.Printf("listening on port:%s:\n", viper.GetString("port"))
	ln, err := net.Listen("tcp", ":"+ viper.GetString("port"))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	msgchan := make(chan string)
	addchan := make(chan Client)
	rmchan := make(chan Client)

	go handleMessages(msgchan, addchan, rmchan)

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}

		go handleConnection(conn, msgchan, addchan, rmchan)
	}
}

func (c Client) ReadLinesInto(ch chan<- string) {
	bufc := bufio.NewReader(c.conn)
	for {
		line, err := bufc.ReadString('\n')
		if err != nil {
			break
		}
		ch <- fmt.Sprintf("%s: %s", c.nickname, line)
	}
}

func (c Client) WriteLinesFrom(ch <-chan string) {
	for msg := range ch {
		_, err := io.WriteString(c.conn, msg)
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

func handleConnection(c net.Conn, msgchan chan<- string, addchan chan<- Client, rmchan chan<- Client) {
	bufc := bufio.NewReader(c)
	defer c.Close()
	client := Client{
		conn:     c,
		nickname: promptNick(c, bufc),
		ch:       make(chan string),
	}
	if strings.TrimSpace(client.nickname) == "" {
		io.WriteString(c, "Invalid Username\n")
		return
	}

	// Register user
	addchan <- client
	defer func() {
		msgchan <- fmt.Sprintf("User %s left the chat room.\n", client.nickname)
		log.Printf("Connection from %v closed.\n", c.RemoteAddr())
		rmchan <- client
	}()
	io.WriteString(c, fmt.Sprintf("Welcome, %s!\n\n", client.nickname))
	msgchan <- fmt.Sprintf("New user %s has joined the chat room.\n", client.nickname)

	// I/O
	go client.ReadLinesInto(msgchan)
	client.WriteLinesFrom(client.ch)
}

func handleMessages(msgchan <-chan string, addchan <-chan Client, rmchan <-chan Client) {
	clients := make(map[net.Conn]chan<- string)

	for {
		select {
		case msg := <-msgchan:
			log.Printf("New message: %s", msg)
			for _, ch := range clients {
				go func(mch chan<- string) { mch <- "\033[1;33;40m" + msg + "\033[m" }(ch)
			}
		case client := <-addchan:
			log.Printf("New client: %v\n", client.conn)
			clients[client.conn] = client.ch
		case client := <-rmchan:
			log.Printf("Client disconnects: %v\n", client.conn)
			delete(clients, client.conn)
		}
	}
}
