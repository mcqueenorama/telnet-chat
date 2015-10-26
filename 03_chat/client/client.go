package client

import (
	m "../message"
	"io"
)

// Client is the struct that encapsulates a connection
//
//    Conn is the connection that's read/written to/from
//    Id is a unique identifier for the connection - the IP as this time
//    Nickname is the chat room user's nick which is associated this message
//    Ch is the channel which is used to talk to the client that sent this message
//    Kind is the kind of connection like a telnet or api or websocket
type Client struct {
	Conn     io.ReadWriteCloser
	Id       string
	Nickname string
	Ch       chan m.ChatMsg
	Kind     string
}
