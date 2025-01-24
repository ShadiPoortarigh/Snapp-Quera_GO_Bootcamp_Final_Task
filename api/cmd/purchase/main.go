package main

import (
	"Snapp-Quera_GO_Bootcamp_Final_Task/api/internal/purchase"
	"Snapp-Quera_GO_Bootcamp_Final_Task/api/pkg"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/nats-io/nats.go"
)

func main() {

	opts := []nats.Option{nats.Name("NATS Sample Replier")}
	opts = setupConnOptions(opts)

	if nc, err := nats.Connect(nats.DefaultURL, opts...); err != nil {
		log.Fatal(err)
	} else {
		pkg.RegisterAdapters(nc)
		listenAndServe(nc)
	}
}

func listenAndServe(nc *nats.Conn) {

	for k, v := range purchase.Handlers {
		nc.QueueSubscribe(k, k, handleMsg(nc, v))
		log.Printf("Listening on [%s]", k)
	}

	nc.Flush()
	if err := nc.LastError(); err != nil {
		log.Fatal(err)
	}
	drainBeforeExit(nc)
}

func handleMsg(nc *nats.Conn, h purchase.Handler) nats.MsgHandler {
	return func(msg *nats.Msg) {
		purchase.MsgCount += 1
		var reply []byte

		reply = h(msg.Data)

		msg.Respond(reply)
		log.Printf("[#] Request [%s]: '%s' was replied '%s'", msg.Subject, string(msg.Data), string(reply))
	}
}

func setupConnOptions(opts []nats.Option) []nats.Option {
	totalWait := 10 * time.Minute
	reconnectDelay := time.Second

	opts = append(opts, nats.ReconnectWait(reconnectDelay))
	opts = append(opts, nats.MaxReconnects(int(totalWait/reconnectDelay)))
	opts = append(opts, nats.DisconnectHandler(func(nc *nats.Conn) {
		log.Printf("Disconnected: will attempt reconnects for %.0fm", totalWait.Minutes())
	}))
	opts = append(opts, nats.ReconnectHandler(func(nc *nats.Conn) {
		log.Printf("Reconnected [%s]", nc.ConnectedUrl())
	}))
	opts = append(opts, nats.ClosedHandler(func(nc *nats.Conn) {
		log.Fatalf("Exiting: %v", nc.LastError())
	}))
	return opts
}

// graceful shutdown
func drainBeforeExit(nc *nats.Conn) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	log.Println()
	log.Printf("Draining...")
	nc.Drain()
	log.Fatalf("Exiting")
}
