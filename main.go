package main

import (
	"fmt"
	"github.com/nikola43/realtimechat/server"
)

func main() {
	fmt.Println("Hello World")
	s := server.New()
	s.Start()
}