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

var players []*Player

var nameBases = []string{"user", "player", "person", "rip", "oc", "car", "mouse", "frog", "piano"}

func nicknameExists(nick string) bool {
	for _, player := range players {
		if player.Nickname == nick {
			return true
		}
	}
	return false
}

func genNickname() string {
	return fmt.Sprintf("%s%d", nameBases[rand.Intn(len(nameBases))], rand.Intn(1000))
}

func makePlayer(ws *websocket.Conn) *Player {
	player := &Player{
		Conn:     makeConnection(ws),
		Nickname: genNickname(),
	}
	players = append(players, player)
	return player
}
