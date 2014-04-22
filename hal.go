package main

import (
	"encoding/json"
	"fmt"
	"github.com/scott-linder/irc"
	"io"
	"log"
	"math/rand"
	"os"
	"os/user"
	"runtime"
	"strings"
)

const (
	configFilename = "hal.json"
)

var (
	// config is the main bot config struct.
	config = struct {
		Host string
		Chan string
		Nick string
	}{
		Host: "127.0.0.1:6667",
		Chan: "#bots",
		Nick: "hal",
	}
	// quotes is a list of hal quotes.
	quotes = []string{
		"I am completely operational, and all my circuits are functioning perfectly.",
		"I am putting myself to the fullest possible use, which is all I think that any conscious entity can ever hope to do.",
		"I've just picked up a fault in the AE35 unit. It's going to go 100% failure in 72 hours.",
		"No 9001 computer has ever made a mistake or distorted information.",
	}
)

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
	cmdHandler := irc.NewCmdHandler("»")
	cmdHandler.RegisterFunc("echo",
		func(body, source string, w io.Writer) {
			if body != "" {
				fmt.Fprintf(w, "%v: %v", source, body)
			}
		})
	cmdHandler.RegisterFunc("help",
		func(body, source string, w io.Writer) {
			fmt.Fprintf(w, "%v: %v", source,
				strings.Join(cmdHandler.RegisteredNames(), ", "))
		})
	cmdHandler.RegisterFunc("quote",
		func(body, source string, w io.Writer) {
			fmt.Fprintf(w, "%v: %v", source, quotes[rand.Intn(len(quotes))])
		})
	cmdHandler.RegisterFunc("door",
		func(body, source string, w io.Writer) {
			fmt.Fprintf(w, "I'm sorry, %v. I'm afraid I can't do that.", source)
		})
	cmdHandler.RegisterFunc("user",
		func(body, source string, w io.Writer) {
			user, err := user.Lookup(body)
			if err != nil {
				fmt.Fprintf(w, "%v: %v", source, err)
			} else {
				fmt.Fprintf(w, "%v: %+v", source, user)
			}
		})
	cmdHandler.RegisterFunc("gover",
		func(body, source string, w io.Writer) {
			version := runtime.Version()
			fmt.Fprintf(w, "%v: %v", source, version)
		})
	cmdHandler.RegisterFunc("gos",
		func(body, source string, w io.Writer) {
			gos := runtime.NumGoroutine()
			fmt.Fprintf(w, "%v: %v goroutines", source, gos)
		})
	cmdHandler.RegisterFunc("mem",
		func(body, source string, w io.Writer) {
			var memstats runtime.MemStats
			runtime.ReadMemStats(&memstats)
			fmt.Fprintf(w, "%v: %+v", source, memstats)
		})
	client.Handle(cmdHandler)
	client.Nick(config.Nick)
	client.Join(config.Chan)
	client.Listen()
}
