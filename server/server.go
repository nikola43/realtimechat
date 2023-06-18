package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"github.com/antoniodipinto/ikisocket"
	"github.com/gofiber/websocket/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

type Server struct {
	http    *fiber.App
	ws     *ikisocket.Websocket
	socketClients map[string]string
}

type SocketEvent struct {
	Type   string      `json:"type"`
	Action string      `json:"action"`
	Data   interface{} `json:"data"`
}

func New() *Server {

	httpServer := fiber.New(fiber.Config{
		BodyLimit: 2000 * 1024 * 1024, // this is the default limit of 4MB
	})

	socketClients := make(map[string]string, 0)

	httpServer.Use(cors.New(cors.Config{
		AllowOrigins: "*",
	}))

	httpServer.Use(logger.New())
	httpServer.Use(cors.New(cors.Config{}))

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

	/*
		ws.Get("/:id", ikisocket.New(func(kws *ikisocket.Websocket) {
			websockets.SocketInstance = kws

			// Retrieve the user id from endpoint
			userId := kws.Params("id")

			// Add the connection to the list of the connected clients
			// The UUID is generated randomly and is the key that allow
			// ikisocket to manage Emit/EmitTo/Broadcast
			websockets.SocketClients[userId] = kws.UUID

			// Every websocket connection has an optional session key => value storage
			kws.SetAttribute("user_id", userId)

			//Broadcast to all the connected users the newcomer
			// kws.Broadcast([]byte(fmt.Sprintf("New user connected: %s and UUID: %s", userId, kws.UUID)), true)
			//Write welcome message
			kws.Emit([]byte(fmt.Sprintf("Socket connected")))
		}))
	*/

	// Pull out in another function
	// all the ikisocket callbacks and listeners
	// Multiple event handling supported
	ikisocket.On(ikisocket.EventConnect, func(ep *ikisocket.EventPayload) {
		fmt.Println(fmt.Sprintf("Connection socket event - User: %s", ep.Kws.GetStringAttribute("user_id")))
	})

	// On message event
	ikisocket.On(ikisocket.EventMessage, func(ep *ikisocket.EventPayload) {
		fmt.Println(fmt.Sprintf("Message socket event - User: %s", ep.Kws.GetStringAttribute("user_id")))
	})

	// On disconnect event
	ikisocket.On(ikisocket.EventDisconnect, func(ep *ikisocket.EventPayload) {
		// Remove the user from the local clients
		delete(socketClients, ep.Kws.GetStringAttribute("user_id"))
		fmt.Println(fmt.Sprintf("Disconnection event - User: %s", ep.Kws.GetStringAttribute("user_id")))
	})

	// On close event
	// This event is called when the server disconnects the user actively with .Close() method
	ikisocket.On(ikisocket.EventClose, func(ep *ikisocket.EventPayload) {
		// Remove the user from the local clients
		delete(socketClients, ep.Kws.GetStringAttribute("user_id"))
		fmt.Println(fmt.Sprintf("Close event - User: %s", ep.Kws.GetStringAttribute("user_id")))
	})

	// On error event
	ikisocket.On(ikisocket.EventError, func(ep *ikisocket.EventPayload) {
		fmt.Println(fmt.Sprintf("Error event - User: %s", ep.Kws.GetStringAttribute("user_id")))
	})

	return &Server{
		http: httpServer,
		ws:  new(ikisocket.Websocket),
	}
}

func WebSocketUpgradeMiddleware(context *fiber.Ctx) error {
	// IsWebSocketUpgrade returns true if the client
	// requested upgrade to the WebSocket protocol.
	if websocket.IsWebSocketUpgrade(context) {
		context.Locals("allowed", true)
		return context.Next()
	}

	return fiber.ErrUpgradeRequired
}

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

func (s *Server) Start() {
	err := s.http.Listen(":3000")
	if err != nil {
		log.Fatal(err)
	}
}

func (s *Server) Stop() {
	fmt.Println("Server stopped")
	s.http.Shutdown()
}
