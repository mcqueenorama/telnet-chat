package message

import (
	"fmt"
	"time"
)

// Channel is the struct that's passed to the message handling dispatch routine
type ChatMsg struct {
	Nick    string
	Channel string
	Msg     string
	Time    string
}

// String is a stringer for ChatMsg
//
// Its the place where the brodcast message format is specified
//
//move this fmt string into config file
//
//found a page showing what this stuff means
//https://xdevs.com/guide/color_serial/
func (c ChatMsg) String() string {
	return fmt.Sprintf("\033[1;33;40m%s: %s \033[m:%s\033[m", c.Time, c.Nick, c.Msg)
}

// MakeChatMessage takes a nick and printf-style arguments to 
// produce a ChatMsg struct
//
// it does not send anything, it just makes the struct
func NewChatMessage(nick string, format string, args ...interface{}) ChatMsg {

	return ChatMsg{Nick: nick, Time: time.Now().Format(time.Kitchen), Msg: fmt.Sprintf(format, args...)}

}
