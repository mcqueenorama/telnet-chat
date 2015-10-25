package client

import (
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