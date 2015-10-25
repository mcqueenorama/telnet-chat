package api

import (
	"fmt"
	"net/http"
	"strings"

	"../client"
	log "../logger"
	m "../message"
)

func ApiServer(port string, msgchan chan m.ChatMsg, addchan chan client.Client, rmchan chan client.Client) {

	http.HandleFunc("/chat/", func(w http.ResponseWriter, req *http.Request) {

		var channel, nick, msg string

		urlParts := strings.Split(req.URL.Path, "/")

		log.Info("api call:%s:parts:%d:\n", req.URL.Path, len(urlParts))

		if len(urlParts) < 5 {
			http.NotFound(w, req)
			return
		} else if urlParts[3] == "" {
			http.NotFound(w, req)
			return
		} else if urlParts[4] == "" {
			http.NotFound(w, req)
			return
		} else {
			channel = urlParts[2]
			nick = urlParts[3]
			msg = strings.Join(urlParts[4:], "/")
		}

		log.Info("api call:channel:%s:nick:%s:msg:%s:\n", channel, nick, msg)

		msgchan <- m.MakeChatMessage(nick, "%s\n", msg)

		fmt.Fprintf(w, "sending message for:user:%s:to chan:%s:\n", nick, channel)

	})

	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}

}
