package main

import (
	"log"
	"strings"

	"github.com/Kong/go-pdk"
	"github.com/Kong/go-pdk/server"
)

type Config struct {
	Attach bool
}

func New() interface{} {
	return &Config{}
}

const Version string = "0.0.1"
const Priority = 1

func main() {
	err := server.StartServer(New, Version, Priority)
	if err != nil {
		log.Fatal(err)
	}
}

func (c Config) Access(kong *pdk.PDK) {
	userAgent, _ := kong.Request.GetHeader("user-agent")
	log.Printf("GO-PLUG: Got request from %s", userAgent)

	if c.Attach {
		if strings.Contains(userAgent, "Kong Builders") {
			err := kong.Response.SetHeader("X-Kong-Builders", "Welcome to the jungle 🌴")
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}
