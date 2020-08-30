package main

import (
	"fmt"

	"github.com/pugovok/goya/app"
)

func main() {
	var server app.Server

	err := server.LoadConfig("config")
	if err != nil {
		panic(err)
	}

	err = server.InitLogger()
	if err != nil {
		fmt.Println(err)
	}

	server.Run()
}
