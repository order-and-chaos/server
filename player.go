package main

import (
	"fmt"
	"math/rand"

	"golang.org/x/net/websocket"
)

// Player holds the state of a player.
type Player struct {
	ID       string
	Nickname string
	Ready    bool
	Conn     *Connection
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
	var nick string

	loop := true
	for loop {
		nick = fmt.Sprintf("%s%d", nameBases[rand.Intn(len(nameBases))], rand.Intn(1000))
		loop = nicknameExists(nick)
	}

	return nick
}

func makePlayer(ws *websocket.Conn) *Player {
	player := &Player{
		ID:       UniqIDf(),
		Nickname: genNickname(),
		Conn:     makeConnection(ws),
	}
	players = append(players, player)
	return player
}
