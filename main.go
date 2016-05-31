package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"golang.org/x/net/websocket"
)

const defaultPort = "1337"

var roomMap = make(map[string]*Room)

func mkRoom() *Room {
	id := UniqIdf()
	room := &Room{
		ID: id,
	}
	roomMap[id] = room
	return room
}

func getRoom(id string) *Room {
	return roomMap[id]
}

func printHelp() {
	fmt.Printf("usage: %s [port=%s]\n", os.Args[0], defaultPort)
}

func main() {
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
