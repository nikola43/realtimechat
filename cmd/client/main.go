package main

import (
	"fmt"

	"github.com/eiannone/keyboard"
	"github.com/nikola43/realtimechat/ws"
)

func main() {
	// connect to websocket server
	serverAddr := "ws://localhost:3000/ws"

	c := ws.Conn{
		OnConnected: func(w *ws.Conn) {
			fmt.Println("Connected")
		},
		OnMessage: func(msg []byte, w *ws.Conn) {
			fmt.Printf("Received message: %s\n", msg)
		},
		OnError: func(err error) {
			fmt.Printf("** ERROR **\n%s\n", err.Error())
		},
	}
	// Connect
	c.Dial(serverAddr, "")
	c.Send(ws.Msg{Body: []byte("Hello World")})
	// Wait for messages

	if err := keyboard.Open(); err != nil {
		panic(err)
	}
	defer func() {
		_ = keyboard.Close()
	}()

	fmt.Println("Press ESC to quit")
	text := ""
	for {
		char, key, err := keyboard.GetKey()
		if err != nil {
			panic(err)
		}
		//fmt.Printf("You pressed: rune %q, key %X\r\n", char, key)
		if key == keyboard.KeyEnter {
			c.Send(ws.Msg{Body: []byte(text)})
			text = ""
		} else {
			text += string(char)
		}

		if key == keyboard.KeyEsc {
			break
		}
	}

}
