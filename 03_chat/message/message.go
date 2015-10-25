package message

import (
"fmt"
"time"
)

type ChatMsg struct {
	Nick string
	Channel   string
	Msg  string
	Time   string
}

//move this fmt string into config file
//
//found a page showing what this stuff means
//https://xdevs.com/guide/color_serial/
func (c ChatMsg) String() string {
	return fmt.Sprintf("\033[1;33;40m%s: %s \033[m:%s\033[m", c.Time, c.Nick, c.Msg)
}

func MakeChatMessage(nick string, format string, args ...interface{}) ChatMsg {

	return ChatMsg{Nick: nick, Time: time.Now().Format(time.Kitchen), Msg: fmt.Sprintf(format, args...)}

}