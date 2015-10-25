package client

import (
"bufio"
"fmt"
"io"
// "log"
"strings"

"../logger"
m "../message"
)

type Client struct {
	Conn     io.ReadWriteCloser
	Id string
	Nickname string
	Ch       chan m.ChatMsg
	Kind string
}


func promptNick(c io.ReadWriter, bufc *bufio.Reader) string {
	io.WriteString(c, "\033[1;30;41mWelcome to the fancy demo chat!\033[0m\n")
	io.WriteString(c, "What is your nick? ")
	nick, _, _ := bufc.ReadLine()
	return string(nick)
}

func (c Client) ReadLinesInto(ch chan m.ChatMsg) {
	bufc := bufio.NewReader(c.Conn)
	for {
		line, err := bufc.ReadString('\n')
		if err != nil {
			break
		}

		ch <- m.MakeChatMessage(c.Nickname, "%s", line)
	}
}

func (c Client) WriteLinesFrom(ch chan m.ChatMsg) {
	for msg := range ch {
		_, err := io.WriteString(c.Conn, msg.String())
		if err != nil {
			return
		}
	}
}

func HandleTelnetConnection(c io.ReadWriteCloser, id string, msgchan chan m.ChatMsg, addchan chan Client, rmchan chan Client, log *logger.Log) {

	bufc := bufio.NewReader(c)
	defer c.Close()
	client := Client{
		Conn:     c,
		Nickname: promptNick(c, bufc),
		Ch:       make(chan m.ChatMsg),
		Id: id,
		Kind: "telnet",
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
	go client.ReadLinesInto(msgchan)
	client.WriteLinesFrom(client.Ch)

}