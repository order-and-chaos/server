package main

import (
	"fmt"
	"math/rand"

	"golang.org/x/net/websocket"
)

// Player holds the state of a player.
type Player struct {
	Conn     *Connection
	Nickname string
}

var nameBases = []string{"user", "player", "person", "rip", "oc", "car", "mouse", "frog", "piano"}

func genNickname() string {
	return fmt.Sprintf("%s%d", nameBases[rand.Intn(len(nameBases))], rand.Intn(1000))
}

func makePlayer(ws *websocket.Conn) *Player {
	return &Player{
		Conn:     makeConnection(ws),
		Nickname: genNickname(),
	}
}
