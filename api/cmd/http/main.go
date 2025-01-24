package main

import (
	api "Snapp-Quera_GO_Bootcamp_Final_Task/api/internal/http"
	"log"
	"net/http"

	"github.com/nats-io/nats.go"
)

func main() {
	nc := connectNats()
	http.HandleFunc("/rate", api.CreateHandlerWithNats(nc, "rate"))
	http.HandleFunc("/purchase", api.CreateHandlerWithNats(nc, "purchase"))
	http.HandleFunc("/sell", api.CreateHandlerWithNats(nc, "sell"))

	log.Printf("Serving on :8090")
	if err := http.ListenAndServe(":8090", nil); err != nil {
		log.Printf("Can't serve on :8090 - %s", err.Error())
	}
}

func connectNats() *nats.Conn {
	log.Printf("connecting to nats")
	opts := []nats.Option{nats.Name("NATS Sample Requestor")}

	if nc, err := nats.Connect(nats.DefaultURL, opts...); err != nil {
		log.Fatal(err)
		return nil
	} else {
		log.Printf("connected to nats")
		return nc
	}
}
