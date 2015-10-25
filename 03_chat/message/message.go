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

func MakeChatMessage(nick string, format string, args ...interface{}) ChatMsg {

	return ChatMsg{Nick: nick, Time: time.Now().Format(time.Kitchen), Msg: fmt.Sprintf(format, args...)}

}