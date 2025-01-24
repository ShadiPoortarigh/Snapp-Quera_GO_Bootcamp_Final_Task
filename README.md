### Snapp-Quera_GO_Bootcamp_Final_Task

In this project, I have implemented the use of NATS in the form of two separate and independent applications that I named them 'chat' and 'api'.

Lets talk about each one separately:

##### chat:
The code implements a real-time chat application using NATS. The application supports features like join and leave chat groups dynamically, broadcast messages to all members of a selected group and view active users in a group.

###### How Components Communicate in Chat:
1. Main Package: 
    Starts the application and sets up the NATS connection.
    Delegates chat functionalities to the ShowChatOnConsole function.

2. Internal Package:
    NATS Communication: Handles publishing and subscribing to topics.
    User Events: Updates and synchronizes user states across clients using the system.users topic.
    Messaging: Publishes user messages to the current group and receives messages from subscribed groups.

3. Shared State:
    Subscriptions (map[string]*nats.Subscription): Tracks groups the user is subscribed to.
    Active Users (map[string]map[string]bool): Tracks users currently active in each group.
    Broadcast Logs: Maintains the last broadcast details for each group.

###### How to Run Chat:
First, we need to pull the NATS server image and run it in a Docker container:

$ docker pull nats
$ docker run --name nats --network nats --rm -p 4222:4222 -p 8222:8222 nats --http_port 8222

Then open a terminal and run the app:
$ go run chat/cmd/main.go

as soon as you enter the command, it asks you for your user name, enter something and press Enter::
$ Maryam

Now you can create a group and join the group. For example:
$ +Golang
$ @Golang

You can open any number of terminal windows and follow the instructions to register and join the group
and start chatting. Whatever you send will be received by all active members of the group in real time.
In addition, if #users command is run in the terminal window of the first member, it will list the name
of all active members of the group.


##### api application:
The second app implements a very simple distributed system that uses NATS for inter-service communication. It provides APIs for rate, purchase and sell services.

###### Communication Flow
1. Client Requests: 
    The client interacts with the system through HTTP endpoints (e.g., /rate, /purchase, /sell).

2. Request Dispatch: 
    HTTP requests are translated into NATS messages using the respective subject (e.g., rate, purchase, sell).

3. Service Handlers: 
    Different microservices listen to these subjects and process the requests.

4. Response Handling: 
    The service handlers send replies back to the requester via NATS.

5. Client Response: 
    Replies from NATS are forwarded back to the HTTP client.

###### Main Components
1. HTTP API Layer (api):
    Serves HTTP endpoints using Go’s http package.
    Routes requests to NATS subjects.
    Handles responses from NATS and returns them to the client.
    
2. NATS Messaging:
    The system uses NATS for communication between services.
    Implements QueueSubscribe for load balancing and resilience.
    
3. Services:
    Rate Service: Handles rate queries.
    Purchase Service: Processes purchase requests.
    Sell Service: Processes sell requests.

4. Adapters:
    RateAdapter abstracts the logic for retrieving exchange rates via NATS.

5. Utilities:
    Connection management, reconnection handling, and graceful shutdown (drainBeforeExit) to ensure robust operations.


###### How to Run api:
First, we need to pull the NATS server image and run it in a Docker container:

$ docker pull nats
$ docker run --name nats --network nats --rm -p 4222:4222 -p 8222:8222 nats --http_port 8222

Then, strat http server and three other services, each in a separate window terminal:

$ go run api/cmd/http/main.go
$ go run api/cmd/rate/main.go
$ go run api/cmd/sell/main.go
$ go run api/cmd/purchase/main.go

Now we can curl HTTP POST requests to the local server running at 127.0.0.1:8090, targeting the /rate,
/sell and /purchase endpoints:

Get the rate for btc
$ curl '127.0.0.1:8090/rate' -H 'content-type: application/json' --data-raw '{"id":"btc"}'

Get the rate for eth
$ curl '127.0.0.1:8090/rate' -H 'content-type: application/json' --data-raw '{"id":"eth"}'

purchase btc
$ curl '127.0.0.1:8090/purchase' -H 'content-type: application/json' --data-raw '{"accountId":"123","currencyId":"btc","amount":100}'

purchase eth
$ curl '127.0.0.1:8090/purchase' -H 'content-type: application/json' --data-raw '{"accountId":"123","currencyId":"eth","amount":15}'

sell eth
$ curl '127.0.0.1:8090/sell' -H 'content-type: application/json' --data-raw '{"accountId":"123","currencyId":"eth","amount":20}'


##### Tests: 
Finally, There are a couple of test cases, located in tests folder for each application.

$ cd chat/tests
$ go test -v


$ cd api/tests
$ go test -v
