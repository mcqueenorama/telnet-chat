// This is where all the telnet related code is kept
//
// It handles connection establishment, including prompting the user
// for a nick
package telnet

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strings"

	"../client"
	log "../logger"
	m "../message"
)

// promptNick talks ugly telnet and gets a nick from the user
func promptNick(c io.ReadWriter, bufc *bufio.Reader) string {
	io.WriteString(c, "\033[1;30;41mWelcome to the fancy demo chat!\033[0m\n")
	io.WriteString(c, "What is your nick? ")
	nick, _, _ := bufc.ReadLine()
	return string(nick)
}

// ReadLinesInto reads from the client and packs it up into a client struct
// and sends it off to the message handler
func ReadLinesInto(c client.Client, ch chan m.ChatMsg) {
	bufc := bufio.NewReader(c.Conn)
	for {
		line, err := bufc.ReadString('\n')
		if err != nil {
			break
		}

		ch <- m.NewChatMessage(c.Nickname, "%s", line)
	}
}

// WriteLinesFrom reads messages from the meesage handler
// and sends them off to the client
func WriteLinesFrom(c client.Client) {
	for msg := range c.Ch {
		_, err := io.WriteString(c.Conn, msg.String())
		if err != nil {
			return
		}
	}
}

// TelnetServer starts up the telnet listener and starts a go routine to handle inclming messages from telnet clients
func TelnetServer(ip, port string, msgchan chan m.ChatMsg, addchan chan client.Client, rmchan chan client.Client) {

	ln, err := net.Listen("tcp", ip + ":" + port)
	if err != nil {
		log.Error("Listener setup error:%v:\n", err)
		os.Exit(1)
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Error("Listener accept error:%v:\n", err)
			continue
		}

		go handleConnection(conn, conn.RemoteAddr().String(), msgchan, addchan, rmchan)
	}
}

// handleConnection is a per-connection go routine that registers the connection
// and starts the go routines that will read/write from the client
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
		msgchan <- m.NewChatMessage("system", "User %s left the chat room.\n", client.Nickname)
		log.Info("Connection from %v closed.\n", id)
		rmchan <- client
	}()
	io.WriteString(c, fmt.Sprintf("Welcome, %s!\n\n", client.Nickname))

	msgchan <- m.NewChatMessage("system", "New user %s has joined the chat room.\n", client.Nickname)

	// I/O
	go ReadLinesInto(client, msgchan)
	WriteLinesFrom(client)

}
