package client

import (
	m "../message"
	"io"
)

type Client struct {
	Conn     io.ReadWriteCloser
	Id       string
	Nickname string
	Ch       chan m.ChatMsg
	Kind     string
}
