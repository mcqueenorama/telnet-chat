// This is a simple telnet chat server swiped largely from 
//
// https://github.com/akrennmair/telnet-chat
//
// The original code is likely hard to recognize at this time.
//
// I've added logging and a config file, and factored everything into sensible namepspaces
// and add a structure providing an internal message format with useful routeable metadata, and unique
// identifiers for each client while the clients may be entering from different protocols, requiring
// different id generation schemes.  See client.go.
//
// A very small but significant change is to the function signature of the message processing
// code, now requiring only a ReadWriteCloser instead of specifying a net.Conn.  This makes it very easy
// add other kinds of connections.
//
// Demonstrating this, I've added an api so simple HTTP GETs can be used to send messages to the chat.
// The same simple message handler is able to handle the new entry point without change. (Though I did
// factor code out of there too.)
//
// The Config file is TOML: https://github.com/toml-lang/toml
//
// I've used viper for command line arguments because it will handle both a config file and cli options
// in a nice way, and it knows how to do toml.
//
// Logging is via https://github.com/op/go-logging since it has decent features, is very configurable,
// and allows specification via config file.
//
// I've added a wrapper for the logging since I didn't like the way it was called.  With the wrapper it is called more
// like glog, which I think is better, but is less configurable.
package main

import (
	"fmt"

	"github.com/spf13/viper"

	"./api"
	"./client"
	log "./logger"
	m "./message"
	"./telnet"
)

const (
	telnetIPWord = "telnetIP"
	telnetPortWord = "telnetPort"
	apiIPWord    = "apiIP"
	apiPortWord    = "apiPort"
)

func main() {

	var configDefault string = "chat"

	viper.SetConfigName(configDefault)
	viper.SetConfigType("toml")
	viper.AddConfigPath("./")
	viper.SetDefault(telnetIPWord, "127.0.0.1")
	viper.SetDefault(telnetPortWord, "6000")
	viper.SetDefault(apiIPWord, "127.0.0.1")
	viper.SetDefault(apiPortWord, "6001")

	err := viper.ReadInConfig()
	if err != nil {
		fmt.Printf("No configuration file found:%s:err:%v: - using defaults\n", configDefault, err)
	}

	logFile := viper.GetString("logFile")

	log.MustSetupLogging(logFile)

	log.Info("listening on:telnet:%s:%s:api:%s:%s:\n", viper.GetString(telnetIPWord), viper.GetString(telnetPortWord), viper.GetString(apiIPWord), viper.GetString(apiPortWord))

	msgchan := make(chan m.ChatMsg)
	addchan := make(chan client.Client)
	rmchan := make(chan client.Client)

	go handleMessages(msgchan, addchan, rmchan)

	go telnet.TelnetServer(viper.GetString(telnetIPWord), viper.GetString(telnetPortWord), msgchan, addchan, rmchan)

	api.ApiServer(viper.GetString(apiIPWord), viper.GetString(apiPortWord), msgchan, addchan, rmchan)

}

func handleMessages(msgchan chan m.ChatMsg, addchan chan client.Client, rmchan chan client.Client) {

	clients := make(map[string]chan m.ChatMsg)

	for {
		select {
		case msg := <-msgchan:
			log.Info("New message: %s", msg.Msg[0:len(msg.Msg) - 1])
			for _, ch := range clients {
				go func(mch chan m.ChatMsg, _msg m.ChatMsg) { mch <- _msg }(ch, msg)
			}

		case client := <-addchan:
			log.Info("New client:id:%v:\n", client.Id)
			clients[client.Id] = client.Ch

		case client := <-rmchan:
			log.Info("Client disconnects: %v\n", client.Id)
			delete(clients, client.Id)
		}
	}

}
