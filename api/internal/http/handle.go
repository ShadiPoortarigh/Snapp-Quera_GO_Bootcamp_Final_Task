package api

import (
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/nats-io/nats.go"
)

var msgCount int

func CreateHandlerWithNats(nc *nats.Conn, subj string) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		msgCount += 1
		if bt, err := ioutil.ReadAll(req.Body); err != nil {
			log.Printf("Error reading body: %v", err)
			http.Error(w, "can't read body", http.StatusBadRequest)
			return
		} else {
			log.Printf("Sending a request [%s]", subj)
			if reply, err := request(nc, subj, bt); err != nil {
				log.Printf("Service reply error [%s] : '%s'", subj, err.Error())
				http.Error(w, "can't process request", http.StatusInternalServerError)
			} else {
				log.Printf("Replied [%s] '%s'", subj, string(reply))
				w.Write(reply)
			}
		}
	}
}

func request(nc *nats.Conn, subj string, req []byte) ([]byte, error) {

	if reply, err := nc.Request(subj, req, 2*time.Second); err != nil {
		if e := nc.LastError(); e != nil {
			return nil, e
		}
		return nil, err
	} else {
		return reply.Data, nil
	}
}
