package main

import (
	"log"

	"github.com/nats-io/nats.go"
)

// publish is the internal function to publish messages to a nats-server.
// Sends a protocol data message by queuing into the bufio writer
// and kicking the flush go routine. These writes should be protected.
func publish(nc *nats.Conn, subj, msg string) {

	nc.Publish(subj, []byte(msg))

	if err := nc.LastError(); err != nil {
		log.Printf("error publishing %s on %s group : %s", msg, subj, err.Error())
	} else {
		//log.Printf("Published [%s] : '%s'\n", subj, msg)
	}
}
