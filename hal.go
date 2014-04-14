package main

import (
	"encoding/json"
	"fmt"
	"github.com/scott-linder/irc"
	"log"
	"os"
)

const (
	configFilename = "hal.json"
)

// config is the main bot config struct.
var config = struct {
	Host string
	Chan string
}{
	Host: "irc.freenode.net:6667",
	Chan: "#bots",
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

func (pong Pong) Accept(msg *irc.Msg) bool { return msg.Cmd == "PING" }
func (pong Pong) Handle(msg *irc.Msg, send chan *irc.Msg) {
	send <- &irc.Msg{Cmd: "PONG", Params: msg.Params}
}

// Echo talks back.
type Echo struct{}

func (echo Echo) Accept(msg *irc.Msg) bool { return msg.Cmd == "PRIVMSG" }
func (echo Echo) Handle(msg *irc.Msg, send chan *irc.Msg) {
	send <- &irc.Msg{Cmd: "PRIVMSG", Params: msg.Params}
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
	client.Register(Pong{})
	client.Register(Echo{})
	client.Nick("hal")
	client.Join(config.Chan)
	client.Listen()
}
