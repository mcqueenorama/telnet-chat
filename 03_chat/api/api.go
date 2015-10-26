// this package allows users to use HTTP GET calls to send messages to the chat rooms with a username
//
// the url format is like /chat/$room/$user/$message
package api

import (
	"fmt"
	"net/http"
	"strings"

	"../client"
	log "../logger"
	m "../message"
)

// ApiServer is an http server for posting message to the chat via GET
//
// it does not have persistent connections and needs no connection handlers
//
// it just makes a chat message struct and sends it on to the message handler
func ApiServer(ip, port string, msgchan chan m.ChatMsg, addchan chan client.Client, rmchan chan client.Client) {

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

		msgchan <- m.NewChatMessage(nick, "%s\n", msg)

		//send reply to the caller
		fmt.Fprintf(w, "sending message for:user:%s:to chan:%s:\n", nick, channel)

	})

	err := http.ListenAndServe(ip + ":" + port, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}

}
