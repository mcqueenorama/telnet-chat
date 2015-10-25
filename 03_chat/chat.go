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
	telnetPortName = "telnetPort"
	apiPortName    = "apiPort"
)

func main() {

	var configDefault string = "chat"

	viper.SetConfigName(configDefault)
	viper.SetConfigType("toml")
	viper.AddConfigPath("./")
	viper.SetDefault(telnetPortName, "6000")
	viper.SetDefault(apiPortName, "6001")

	err := viper.ReadInConfig()
	if err != nil {
		fmt.Printf("No configuration file found:%s:err:%v: - using defaults\n", configDefault, err)
	}

	logFile := viper.GetString("logFile")

	log.MustSetupLogging(logFile)

	log.Info("listening on ports:telnet:%s:api:%s:\n", viper.GetString(telnetPortName), viper.GetString(apiPortName))

	msgchan := make(chan m.ChatMsg)
	addchan := make(chan client.Client)
	rmchan := make(chan client.Client)

	go handleMessages(msgchan, addchan, rmchan)

	go telnet.TelnetServer(viper.GetString(telnetPortName), msgchan, addchan, rmchan)

	api.ApiServer(viper.GetString(apiPortName), msgchan, addchan, rmchan)

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
