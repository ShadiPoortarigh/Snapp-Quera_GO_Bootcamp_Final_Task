package internal

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/nats-io/nats.go"
)

var cReset = "\033[0m"
var cRed = "\033[31m"
var cGreen = "\033[32m"
var cYellow = "\033[33m"
var cBlue = "\033[34m"
var cPurple = "\033[35m"
var cCyan = "\033[36m"
var cGray = "\033[37m"
var cWhite = "\033[97m"

type Msg struct {
	User    string
	Message string
	Time    string
}

func (m Msg) ToString() string {
	if bt, err := json.Marshal(m); err != nil {
		log.Printf("cant marshal msg %v", m)
		return ""
	} else {
		return string(bt)
	}
}

func printHelp(user string) {
	fmt.Println("- - - - - - - - - - - - - - - - - - - - - - - - - - - -")
	fmt.Println(cCyan + "hi " + cBlue + user)
	fmt.Println(cYellow + "   +group " + cCyan + "- to be added to a group. e.g. +family")
	fmt.Println(cYellow + "   -group " + cCyan + "- to be removed from a group. e.g. -covid")
	fmt.Println(cYellow + "   @group " + cCyan + "- to select current group. e.g. @meetup")
	fmt.Println(cYellow + "   --help " + cCyan + "-to display this menu")
	fmt.Println(cYellow + "   anything else" + cCyan + " will be published on you current group")
	fmt.Println("- - - - - - - - - - - - - - - - - - - - - - - - - - - -")
}

func wrapSystemMsg(m string) string {
	return cCyan + m + cReset
}

func getUser(reader *bufio.Reader) string {
	fmt.Println(wrapSystemMsg("welcome to the meetup chat"))
	fmt.Println(wrapSystemMsg("*****type your name please****"))
	user, _ := reader.ReadString('\n')
	return strings.Replace(user, "\n", "", -1) //remove CRLF
}

type BroadcastLog struct {
	Sender     string
	Recipients []string
	Group      string
}

var lastBroadcastLog = make(map[string]BroadcastLog)

func ShowChatOnConsole(nc *nats.Conn) {
	reader := bufio.NewReader(os.Stdin)
	user := getUser(reader)
	printHelp(user)

	var curGroup string
	subs := make(map[string]*nats.Subscription)
	activeUsers := make(map[string]map[string]bool)

	// Subscribe to system.users for synchronization
	nc.Subscribe("system.users", func(msg *nats.Msg) {
		var update struct {
			Group  string
			User   string
			Action string // join, leave, switch
		}
		if err := json.Unmarshal(msg.Data, &update); err != nil {
			log.Printf("Error decoding user update: %s", err)
			return
		}

		if activeUsers[update.Group] == nil {
			activeUsers[update.Group] = make(map[string]bool)
		}

		switch update.Action {
		case "join":
			activeUsers[update.Group][update.User] = true
		case "leave":
			delete(activeUsers[update.Group], update.User)
		case "switch":
			for group := range activeUsers {
				delete(activeUsers[group], update.User)
			}
			activeUsers[update.Group][update.User] = true
		}
	})

	// Broadcast user events
	broadcastUserUpdate := func(group, user, action string) {
		update := struct {
			Group  string
			User   string
			Action string
		}{
			Group:  group,
			User:   user,
			Action: action,
		}
		if data, err := json.Marshal(update); err == nil {
			nc.Publish("system.users", data)
		} else {
			log.Printf("Error encoding user update: %s", err)
		}
	}

	prompt := func() {
		fmt.Print(cReset + curGroup + "$ ")
	}

	printMsg := func(s []byte, group string) {
		var m Msg
		if err := json.Unmarshal([]byte(s), &m); err != nil {
			log.Printf("Can't unmarshal msg %s", s)
		} else if m.User != user {
			fmt.Printf(cPurple+"\n%s:(%s):%s\n"+cReset, group, m.User, m.Message)
			prompt()
		}
	}

	for {
		prompt()
		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(text)

		if len(text) == 0 {
			continue
		}

		if text == "--help" {
			printHelp(user)
			continue
		}

		content := text[1:]

		switch text[:1] {
		case "+":
			if ok, err := regexp.MatchString("^[a-zA-Z0-9_.-]*$", content); !ok || err != nil {
				fmt.Printf("Group name can only use letters, numbers, and dashes\n")
			} else if s, err := Subscribe(nc, content, printMsg); err != nil {
				fmt.Printf("Can't join group %s : %s\n", content, err)
			} else {
				subs[content] = s
				if activeUsers[content] == nil {
					activeUsers[content] = make(map[string]bool)
				}
				activeUsers[content][user] = true
				broadcastUserUpdate(content, user, "join")
			}
		case "-":
			if s, ok := subs[content]; !ok {
				fmt.Printf("Not registered to group %s\n", content)
			} else {
				s.Unsubscribe()
				delete(subs, content)
				delete(activeUsers[content], user)
				broadcastUserUpdate(content, user, "leave")
				fmt.Printf("I'm not in the %s group anymore :(\n", content)
				if content == curGroup {
					curGroup = ""
				}
			}
		case "@":
			if _, ok := subs[content]; !ok {
				fmt.Printf("You are not registered to group %s\n", content)
			} else {
				// Broadcast leaving the current group
				if curGroup != "" {
					broadcastUserUpdate(curGroup, user, "leave")
				}

				// Switch to the new group
				curGroup = content
				broadcastUserUpdate(curGroup, user, "switch")
			}
		case "#":
			if content == "users" {
				if curGroup == "" {
					fmt.Println("You must first select a group, type @family for example")
				} else {
					fmt.Printf("Active users in group '%s':\n", curGroup)
					for u := range activeUsers[curGroup] {
						fmt.Println("-", u)
					}
					if log, ok := lastBroadcastLog[curGroup]; ok {
						fmt.Printf("\nLast broadcast in group '%s':\n", curGroup)
						fmt.Println("Sender:", log.Sender)
						if len(log.Recipients) > 0 {
							fmt.Println("Recipients:")
							for _, recipient := range log.Recipients {
								fmt.Println("-", recipient)
							}
						} else {
							fmt.Println("")
						}
					} else {
						fmt.Println("")
					}
				}
			}

		default:
			if curGroup == "" {
				fmt.Println("You must first select a group, type @family for example")
			} else {
				m := Msg{User: user, Message: text, Time: time.Now().Format("15:06")}
				Publish(nc, curGroup, m.ToString())

				var recipients []string
				for u := range activeUsers[curGroup] {
					if u != user {
						recipients = append(recipients, u)
					}
				}

				lastBroadcastLog[curGroup] = BroadcastLog{
					Sender:     user,
					Recipients: recipients,
					Group:      curGroup,
				}
			}

		}
	}
}
