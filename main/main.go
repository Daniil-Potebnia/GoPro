package main

import (
	"first_project/agent"
	"first_project/server"
)

func main() {
	main := agent.CreateMain()
	main.Start()
	server.StartServer()
}
