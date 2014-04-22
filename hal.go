package main

import (
	"encoding/json"
	"fmt"
	"github.com/scott-linder/irc"
	"log"
	"os"
	"strings"
)

const (
	configFilename = "hal.json"
)

// config is the main bot config struct.
var config = struct {
	Host string
	Chan string
	Nick string
}{
	Host: "127.0.0.1:6667",
	Chan: "#bots",
	Nick: "hal",
}

// loadConfig attempts to load config from configFilename.
func loadConfig() error {
	file, err := os.Open(configFilename)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return err
	}
	return nil
}

// Pong plays your game.
type Pong struct{}

func (pong Pong) Accepts(msg *irc.Msg) bool { return msg.Cmd == "PING" }
func (pong Pong) Handle(msg *irc.Msg, send chan<- *irc.Msg) {
	send <- &irc.Msg{Cmd: "PONG", Params: msg.Params}
}

// Open the pod bay doors, hal.
type Open struct{}

func (open Open) Accepts(msg *irc.Msg) bool { return msg.Cmd == "PRIVMSG" }
func (open Open) Handle(msg *irc.Msg, send chan<- *irc.Msg) {
	receiver, body, err := msg.ExtractPrivmsg()
	if err != nil {
		log.Println(err)
		return
	}
	nick, err := msg.ExtractNick()
	if err != nil {
		log.Println(err)
		return
	}
	lowerbody := strings.ToLower(body)
	if strings.Contains(lowerbody, "open the pod bay doors") &&
		strings.Contains(lowerbody, "hal") {

		response := fmt.Sprintf("I can't let you do that, %v.", nick)
		params := []string{receiver, response}
		send <- &irc.Msg{Cmd: "PRIVMSG", Params: params}
	}
}

func main() {
	fmt.Println("I am a HAL 9001 computer.")
	if err := loadConfig(); err != nil {
		log.Printf("error loading config file %v: %v", configFilename, err)
	}
	client, err := irc.Dial(config.Host)
	if err != nil {
		log.Fatal(err)

	}
	client.Handle(Pong{})
	client.Handle(Open{})
	cmdHandler := irc.NewCmdHandler("!")
	cmdHandler.RegisterFunc("echo",
		func(body, source string, w irc.CmdResponseWriter) {
			if body != "" {
				w.Write([]byte(source + ": " + body))
			}
		})
	client.Handle(cmdHandler)
	client.Nick(config.Nick)
	client.Join(config.Chan)
	client.Listen()
}
