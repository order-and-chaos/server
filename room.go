package main

import "math/rand"

type Room struct {
	ID               string
	PlayerA, PlayerB *Player
	RoleA            GameRole
	Spectators       []*Player
	Board            *Board
}

// StartGame starts the game of this room.
func (r *Room) StartGame() bool {
	if r.Board != nil {
		return true
	}

	if r.PlayerA == nil || r.PlayerB == nil {
		return false
	}

	r.SendAll("startgame")

	if rand.Intn(2) == 0 {
		r.RoleA = Order
	} else {
		r.RoleA = Chaos
	}
	r.Board = MakeBoard(Order)

	return true
}

func (r *Room) StopGame() bool {
	if r.Board == nil {
		return true
	}

	r.Board = nil
	r.SendAll("stopgame")

	return true
}

func (r *Room) SendAll(typ string, args ...string) {
	r.PlayerA.Conn.Send(typ, args...)
	r.PlayerB.Conn.Send(typ, args...)
	r.SendSpectators(typ, args...)
}

func (r *Room) SendSpectators(typ string, args ...string) {
	for _, spec := range r.Spectators {
		spec.Conn.Send(typ, args...)
	}
}
