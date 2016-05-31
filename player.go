package main

import "golang.org/x/net/websocket"

type Player struct {
	Conn     *Connection
	Nickname string
}

func makePlayer(ws *websocket.Conn) *Player {
	return &Player{
		Conn: makeConnection(ws),
	}
}
