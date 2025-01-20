// The implementation includes:

// Joining the Chatroom: Users announce their presence by publishing their username to
// the user list channel.
// Real-time Messaging: The chatRoom channel broadcasts messages to all connected users.
// Active User List: A #users command displays the current active users, dynamically updated
// as users join or leave.
// This program uses nats.go for interacting with NATS and provides a simple CLI for interaction.

package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/nats-io/nats.go"
)

const (
	natsURL  = "nats://localhost:4222"
	chatRoom = "chatroom"
	userList = "userlist"
)

var (
	username string
	conn     *nats.Conn
	mu       sync.Mutex
	users    = make(map[string]bool)
)

func main() {
	var err error
	conn, err = nats.Connect(natsURL)
	if err != nil {
		fmt.Println("Error connecting to NATS server:", err)
		return
	}
	defer conn.Close()

	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter your username: ")
	username, _ = reader.ReadString('\n')
	username = strings.TrimSpace(username)
	if username == "" {
		fmt.Println("Username cannot be empty.")
		return
	}

	// Announce user join
	conn.Publish(userList, []byte(username+":join"))

	// Subscribe to messages
	_, err = conn.Subscribe(chatRoom, func(msg *nats.Msg) {
		fmt.Println(string(msg.Data))
	})
	if err != nil {
		fmt.Println("Error subscribing to chatroom:", err)
		return
	}

	// Subscribe to user list updates
	_, err = conn.Subscribe(userList, func(msg *nats.Msg) {
		updateUsers(string(msg.Data))
	})
	if err != nil {
		fmt.Println("Error subscribing to user list:", err)
		return
	}

	fmt.Println("Welcome to the chatroom! Type your messages below.")
	fmt.Println("Commands: \n #users - View active users \n /exit - Leave the chatroom")

	for {
		fmt.Print("Message: ")
		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(text)

		if text == "#users" {
			showActiveUsers()
			continue
		} else if text == "/exit" {
			conn.Publish(userList, []byte(username+":leave"))
			fmt.Println("Goodbye!")
			break
		}

		message := fmt.Sprintf("%s: %s", username, text)
		conn.Publish(chatRoom, []byte(message))
	}
}

func updateUsers(update string) {
	parts := strings.Split(update, ":")
	if len(parts) != 2 {
		return
	}
	name, action := parts[0], parts[1]

	mu.Lock()
	defer mu.Unlock()

	if action == "join" {
		users[name] = true
	} else if action == "leave" {
		delete(users, name)
	}
}

func showActiveUsers() {
	mu.Lock()
	defer mu.Unlock()

	fmt.Println("Active users:")
	for user := range users {
		fmt.Println("-", user)
	}
}
