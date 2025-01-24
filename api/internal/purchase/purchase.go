package purchase

import (
	"Snapp-Quera_GO_Bootcamp_Final_Task/api/pkg"
	"encoding/json"
	"fmt"
	"log"
)

var MsgCount int

type Handler func([]byte) []byte

var Handlers = make(map[string]Handler)

var SvcError = []byte(`{"error" : "internal service error"}`)

func init() {
	Handlers["purchase"] = processPurchase
}

func processPurchase(data []byte) []byte {
	type req struct {
		AccountId  string  `json:"accountId"`
		CurrencyId string  `json:"currencyId"`
		Amount     float64 `json:"amount"`
	}

	type rep struct {
		TrsId string  `json:"tsrId"`
		Rate  float64 `json:"rate"`
		Total float64 `json:"total"`
		Error string  `json:"error,omitempty"`
	}

	var rq req
	var rp rep

	if err := json.Unmarshal(data, &rq); err != nil {
		log.Printf("[#purchase] cant Unmarshal [%s] to request", string(data))
	} else if rate, err := pkg.RateAdapter(rq.CurrencyId); err != nil {
		log.Printf("[#purchase] error getting rate for [%+v] : %s", rq, err.Error())
		return SvcError
	} else {
		rp = rep{TrsId: fmt.Sprint(10000 + MsgCount), Rate: rate, Total: -rate * rq.Amount}
	}
	if reply, err := json.Marshal(rp); err != nil {
		log.Printf("[#purchase] cant Marshal [%+v] to response: %s", rp, err.Error())
		return SvcError
	} else {
		return reply
	}
}
