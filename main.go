package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"golang.org/x/net/websocket"
)

const defaultPort = "1337"

func getRoom(id string) *Room {
	return roomMap[id]
}

func printHelp() {
	fmt.Printf("usage: %s [port=%s]\n", os.Args[0], defaultPort)
}

func main() {
	rand.Seed(time.Now().UnixNano())

	port := defaultPort

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "-h", "help", "--help":
			printHelp()
			return
		default:
			port = os.Args[1]
		}
	}

	http.Handle("/ws", websocket.Handler(WsHandler))
	log.Print("listening on " + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
