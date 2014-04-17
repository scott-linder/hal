package main

import (
	"encoding/json"
	"fmt"
	"github.com/scott-linder/irc"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
)

const (
	configFilename = "hal.json"
)

// config is the main bot config struct.
var config = struct {
	Host string
	Chan string
}{
	Host: "127.0.0.1:6667",
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
func (pong Pong) Handle(msg *irc.Msg, send chan<- *irc.Msg) {
	send <- &irc.Msg{Cmd: "PONG", Params: msg.Params}
}

// Echo talks back.
type Echo struct{}

func (echo Echo) Accept(msg *irc.Msg) bool { return msg.Cmd == "PRIVMSG" }
func (echo Echo) Handle(msg *irc.Msg, send chan<- *irc.Msg) {
	source, body, err := msg.ExtractPrivmsg()
	if err != nil {
		log.Println(err)
		return
	}
	if strings.HasPrefix(body, "!echo ") {
		params := []string{source, strings.TrimPrefix(body, "!echo ")}
		send <- &irc.Msg{Cmd: "PRIVMSG", Params: params}
	}
}

// Words counts words.
type Words struct {
	count      *int
	countMutex *sync.Mutex
}

func NewWords() Words {
	return Words{count: new(int), countMutex: new(sync.Mutex)}
}
func (words Words) Accept(msg *irc.Msg) bool { return msg.Cmd == "PRIVMSG" }
func (words Words) Handle(msg *irc.Msg, send chan<- *irc.Msg) {
	source, body, err := msg.ExtractPrivmsg()
	if err != nil {
		log.Println(err)
		return
	}
	if body == "!words" {
		words.countMutex.Lock()
		params := []string{source, strconv.Itoa(*words.count)}
		words.countMutex.Unlock()
		send <- &irc.Msg{Cmd: "PRIVMSG", Params: params}
	} else {
		words.countMutex.Lock()
		for _, _ = range strings.Fields(body) {
			*words.count += 1
		}
		words.countMutex.Unlock()
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
	client.Register(Pong{})
	client.Register(Echo{})
	client.Register(NewWords())
	client.Nick("hal")
	client.Join(config.Chan)
	client.Listen()
}
