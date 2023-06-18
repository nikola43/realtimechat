package server

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/antoniodipinto/ikisocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/websocket/v2"
)

type Server struct {
	http    *fiber.App
	ws      *ikisocket.Websocket
	clients map[string]string
	rooms   map[string]*Room
}

// MessageObject Basic chat message object
type MessageObject struct {
	Data string `json:"data"`
	From string `json:"from"`
	Room string `json:"room"`
	To   string `json:"to"`
}

// Room Chat Room message object
type Room struct {
	Name  string   `json:"name"`
	UUID  string   `json:"uuid"`
	Users []string `json:"users"`
}

func New() *Server {
	clients := make(map[string]string, 0)
	rooms := make(map[string]*Room, 0)

	httpServer := fiber.New(fiber.Config{
		BodyLimit: 1 * 1024 * 1024, // this is the default limit of 1MB
	})
	httpServer.Use(cors.New(cors.Config{AllowOrigins: "*"}))
	httpServer.Use(logger.New())

	ws := httpServer.Group("/ws")

	// Setup the middleware to retrieve the data sent in first GET request
	ws.Use(func(c *fiber.Ctx) error {
		// IsWebSocketUpgrade returns true if the client
		// requested upgrade to the WebSocket protocol.
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	// On connection event
	ikisocket.On(ikisocket.EventConnect, func(ep *ikisocket.EventPayload) {
		fmt.Println(fmt.Printf("Connection event - User: %s", ep.Kws.GetStringAttribute("user_id")))
	})

	// On message event
	ikisocket.On(ikisocket.EventMessage, func(ep *ikisocket.EventPayload) {
		fmt.Println(fmt.Printf("Message event - User: %s - Message: %s", ep.Kws.GetStringAttribute("user_id"), string(ep.Data)))

		message := MessageObject{}

		// Unmarshal the json message
		// {
		//  "from": "<user-id>",
		//  "to": "<recipient-user-id>",
		//  "room": "<room-id>",
		//  "data": "hello"
		//}
		err := json.Unmarshal(ep.Data, &message)
		if err != nil {
			fmt.Println(err)
			return
		}

		// If the user is trying to send message
		// into a specific group, iterate over the
		// group user socket UUIDs
		if message.Room != "" {
			// Emit the message to all the room participants
			// iterating on all the uuids
			for _, userId := range rooms[message.Room].Users {
				_ = ep.Kws.EmitTo(clients[userId], ep.Data, ikisocket.TextMessage)
			}

			// Other way can be used EmitToList method
			// if you have a []string of ikisocket uuids
			//
			// ep.Kws.EmitToList(list, data)
			//
			return
		}

		// Emit the message directly to specified user
		err = ep.Kws.EmitTo(clients[message.To], ep.Data, ikisocket.TextMessage)
		if err != nil {
			fmt.Println(err)
		}
	})

	// On disconnect event
	ikisocket.On(ikisocket.EventDisconnect, func(ep *ikisocket.EventPayload) {
		// Remove the user from the local clients
		delete(clients, ep.Kws.GetStringAttribute("user_id"))
		fmt.Println(fmt.Printf("Disconnection event - User: %s", ep.Kws.GetStringAttribute("user_id")))
	})

	// On close event
	// This event is called when the server disconnects the user actively with .Close() method
	ikisocket.On(ikisocket.EventClose, func(ep *ikisocket.EventPayload) {
		// Remove the user from the local clients
		delete(clients, ep.Kws.GetStringAttribute("user_id"))
		fmt.Println(fmt.Printf("Close event - User: %s", ep.Kws.GetStringAttribute("user_id")))
	})

	// On error event
	ikisocket.On(ikisocket.EventError, func(ep *ikisocket.EventPayload) {
		fmt.Println(fmt.Printf("Error event - User: %s", ep.Kws.GetStringAttribute("user_id")))
	})

	httpServer.Get("/ws", ikisocket.New(func(kws *ikisocket.Websocket) {
		// Retrieve the user id from endpoint
		//userId := kws.Params("id")
		userId := RandId()

		// Add the connection to the list of the connected clients
		// The UUID is generated randomly and is the key that allow
		// ikisocket to manage Emit/EmitTo/Broadcast
		clients[userId] = kws.UUID

		// Every websocket connection has an optional session key => value storage
		kws.SetAttribute("user_id", userId)

		//Broadcast to all the connected users the newcomer
		kws.Broadcast([]byte(fmt.Sprintf("New user connected: %s and UUID: %s", userId, kws.UUID)), true, ikisocket.TextMessage)
		//Write welcome message
		kws.Emit([]byte(fmt.Sprintf("Hello user: %s with UUID: %s", userId, kws.UUID)), ikisocket.TextMessage)
	}))

	return &Server{
		http:    httpServer,
		ws:      new(ikisocket.Websocket),
		clients: clients,
		rooms:   rooms,
	}
}


func RandId() string {
	length := 8
	seed := rand.New(rand.NewSource(time.Now().UnixNano()))
	charset := "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789abcdefghijklmnopqrstuvwxyz"

	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seed.Intn(len(charset))]
	}

	return string(b)
}

/*
func (s *Server) Emit(socketEvent SocketEvent, id uint) {
	var socketClientId = strconv.FormatUint(uint64(id), 10)
	if uuid, found := s.socketClients[socketClientId]; found {
		event, err := json.Marshal(socketEvent)
		if err != nil {
			fmt.Println(err)
		}

		emitSocketErr := s.ws.EmitTo(uuid, event)
		if emitSocketErr != nil {
			fmt.Println(emitSocketErr)
		}
	}
}
*/

func (s *Server) Start() {
	err := s.http.Listen(":3000")
	if err != nil {
		log.Fatal(err)
	}
}

func (s *Server) Stop() {
	fmt.Println("Server stopped")
	err :=  s.http.Shutdown()
	if err != nil {
		log.Fatal(err)
	}
}
