package telnet

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strings"

	"../client"
	log "../logger"
	m "../message"
)

func promptNick(c io.ReadWriter, bufc *bufio.Reader) string {
	io.WriteString(c, "\033[1;30;41mWelcome to the fancy demo chat!\033[0m\n")
	io.WriteString(c, "What is your nick? ")
	nick, _, _ := bufc.ReadLine()
	return string(nick)
}

func ReadLinesInto(c client.Client, ch chan m.ChatMsg) {
	bufc := bufio.NewReader(c.Conn)
	for {
		line, err := bufc.ReadString('\n')
		if err != nil {
			break
		}

		ch <- m.MakeChatMessage(c.Nickname, "%s", line)
	}
}

func WriteLinesFrom(c client.Client) {
	for msg := range c.Ch {
		_, err := io.WriteString(c.Conn, msg.String())
		if err != nil {
			return
		}
	}
}

func TelnetServer(ln net.Listener, msgchan chan m.ChatMsg, addchan chan client.Client, rmchan chan client.Client) {

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Error("Listener accept error:%v:\n", err)
			continue
		}

		go handleConnection(conn, conn.RemoteAddr().String(), msgchan, addchan, rmchan)
	}
}

func handleConnection(c io.ReadWriteCloser, id string, msgchan chan m.ChatMsg, addchan chan client.Client, rmchan chan client.Client) {

	bufc := bufio.NewReader(c)
	defer c.Close()
	client := client.Client{
		Conn:     c,
		Nickname: promptNick(c, bufc),
		Ch:       make(chan m.ChatMsg),
		Id:       id,
		Kind:     "telnet",
	}

	if strings.TrimSpace(client.Nickname) == "" {
		io.WriteString(c, "Invalid Username\n")
		return
	}

	// Register user
	addchan <- client
	defer func() {
		msgchan <- m.MakeChatMessage("system", "User %s left the chat room.\n", client.Nickname)
		log.Info("Connection from %v closed.\n", id)
		rmchan <- client
	}()
	io.WriteString(c, fmt.Sprintf("Welcome, %s!\n\n", client.Nickname))

	msgchan <- m.MakeChatMessage("system", "New user %s has joined the chat room.\n", client.Nickname)

	// I/O
	go ReadLinesInto(client, msgchan)
	WriteLinesFrom(client)

}
