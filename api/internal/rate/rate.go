package rate

import (
	"encoding/json"
	"log"
)

var MsgCount int

type Handler func([]byte) []byte

var Handlers = make(map[string]Handler)

var SvcError = []byte(`{"error" : "internal service error"}`)

func init() {
	Handlers["rate"] = processRate
}

func processRate(data []byte) []byte {
	type req struct {
		ID string `json:"id"`
	}

	type rep struct {
		Id    string  `json:"id"`
		Rate  float64 `json:"rate"`
		Error string  `json:"error,omitempty"`
	}

	var rq req
	var rp rep

	if err := json.Unmarshal(data, &rq); err != nil {
		log.Printf("[#rate] cant Unmarshal [%s] to request", string(data))
	} else {
		switch rq.ID {
		case "btc":
			rp = rep{Id: rq.ID, Rate: 44452.60}
		case "eth":
			rp = rep{Id: rq.ID, Rate: 7543.21}
		default:
			rp = rep{Id: rq.ID, Error: "currency does not exist!"}
		}
	}

	if reply, err := json.Marshal(rp); err != nil {
		log.Printf("[#rate] cant Marshal [%v] to response: %s", rp, err.Error())
		return SvcError
	} else {
		return reply
	}

}
