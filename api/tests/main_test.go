package main

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	handle "Snapp-Quera_GO_Bootcamp_Final_Task/api/internal/http"

	"github.com/nats-io/nats.go"
)

func TestRateHandler(t *testing.T) {

	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		t.Fatalf("Failed to connect to NATS: %v", err)
	}
	defer nc.Close()

	handler := handle.CreateHandlerWithNats(nc, "rate")

	req := httptest.NewRequest("POST", "/rate", bytes.NewBuffer([]byte(`{"id":"btc"}`)))
	w := httptest.NewRecorder()

	handler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK, got %v", resp.Status)
	}
}

type NatsRequester interface {
	Request(subj string, data []byte, timeout time.Duration) (*nats.Msg, error)
}

type mockNatsConn struct{}

func (m *mockNatsConn) Request(subj string, data []byte, timeout time.Duration) (*nats.Msg, error) {
	response := `{"tsrId":"10001", "rate":7543.21, "total":-15086.42}`
	return &nats.Msg{Data: []byte(response)}, nil
}

func CreateHandlerWithNats(nc NatsRequester, subj string) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {

		bt, err := ioutil.ReadAll(req.Body)
		if err != nil {
			http.Error(w, "can't read body", http.StatusBadRequest)
			return
		}
		defer req.Body.Close()

		reply, err := nc.Request(subj, bt, 2*time.Second)
		if err != nil {
			http.Error(w, "can't process request", http.StatusInternalServerError)
			return
		}

		w.Write(reply.Data)
	}
}

func TestPurchaseHandler(t *testing.T) {
	mockNats := &mockNatsConn{}
	handler := CreateHandlerWithNats(mockNats, "purchase")

	requestBody := `{"accountId":"12345","currencyId":"eth","amount":2}`
	req := httptest.NewRequest("POST", "/purchase", bytes.NewBuffer([]byte(requestBody)))
	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	handler(recorder, req)

	res := recorder.Result()
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK, got %v", res.Status)
	}
	expectedBody := `{"tsrId":"10001", "rate":7543.21, "total":-15086.42}`
	body := recorder.Body.String()
	if body != expectedBody {
		t.Errorf("Expected body %v, got %v", expectedBody, body)
	}
}
