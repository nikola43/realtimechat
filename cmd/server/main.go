package main

import (
	"github.com/nikola43/realtimechat/server"
)

func main() {
	s := server.New()
	s.Start()
}
