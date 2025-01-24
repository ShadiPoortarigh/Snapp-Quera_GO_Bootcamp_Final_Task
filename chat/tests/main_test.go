package main

import (
	"encoding/json"
	"testing"
	"time"

	set "Snapp-Quera_GO_Bootcamp_Final_Task/chat/internal"

	"github.com/nats-io/nats.go"
)

func TestToString(t *testing.T) {
	m := set.Msg{User: "test_user", Message: "Hello, World!", Time: "10:00"}
	result := m.ToString()
	var parsedMsg set.Msg
	if err := json.Unmarshal([]byte(result), &parsedMsg); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %s", err)
	}
	if parsedMsg != m {
		t.Errorf("Expected %v, got %v", m, parsedMsg)
	}
}

func TestSetupConnOptions(t *testing.T) {
	opts := []nats.Option{}
	opts = set.SetupConnOptions(opts)

	if len(opts) == 0 {
		t.Error("Expected options to be set, but got empty")
	}
}

func TestSubscribeAndPublish(t *testing.T) {
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		t.Fatalf("Failed to connect to NATS: %s", err)
	}
	defer nc.Close()

	received := make(chan string, 1)
	sub, err := set.Subscribe(nc, "test_group", func(data []byte, subject string) {
		received <- string(data)
	})
	if err != nil {
		t.Fatalf("Failed to subscribe: %s", err)
	}
	defer sub.Unsubscribe()

	msg := "Hello, NATS!"
	set.Publish(nc, "test_group", msg)

	select {
	case res := <-received:
		if res != msg {
			t.Errorf("Expected %s, got %s", msg, res)
		}
	case <-time.After(2 * time.Second):
		t.Error("Did not receive the message in time")
	}
}
