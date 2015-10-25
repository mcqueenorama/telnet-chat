package client

import (
"bufio"
"io"

m "../message"
)

type Client struct {
	Conn     io.ReadWriteCloser
	Id string
	Nickname string
	Ch       chan m.ChatMsg
	Kind string
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