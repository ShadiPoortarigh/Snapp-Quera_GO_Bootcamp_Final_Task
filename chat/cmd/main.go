package main

import (
	"log"

	set "Snapp-Quera_GO_Bootcamp_Final_Task/chat/internal"

	"github.com/nats-io/nats.go"
)

func main() {

	opts := []nats.Option{nats.Name("Simple chat")}
	opts = set.SetupConnOptions(opts)

	if nc, err := nats.Connect(nats.DefaultURL, opts...); err != nil {
		log.Fatalf("connect error:%s", err.Error())
	} else {
		defer nc.Close()

		set.ShowChatOnConsole(nc)

		nc.Flush()
	}

}
