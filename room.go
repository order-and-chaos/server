package main

import (
	"errors"
	"math/rand"
)

// Room contains the state of a game room.
type Room struct {
	ID               string
	PlayerA, PlayerB *Player
	RoleA, RoleB     GameRole
	Spectators       []*Player
	Board            *Board
}

var roomMap map[string]*Room

func makeRoom() *Room {
	id := UniqIdf()
	room := &Room{
		ID: id,
	}
	roomMap[id] = room
	return room
}

// Started returns if the game for this room has been started or not.
func (r *Room) Started() bool {
	return r.Board != nil
}

// StartGame starts the game of this room.
func (r *Room) StartGame() (started bool) {
	if r.Started() {
		return true
	}

	if r.PlayerA == nil || r.PlayerB == nil {
		return false
	}

	if rand.Intn(2) == 0 {
		r.RoleA = Order
		r.RoleB = Chaos
	} else {
		r.RoleA = Chaos
		r.RoleB = Chaos
	}
	r.Board = MakeBoard(Order)

	r.SendAll("startgame", r.RoleA.String(), r.RoleB.String())

	return true
}

// StopGame stops the game of this room.
func (r *Room) StopGame() bool {
	if !r.Started() {
		return true
	}

	r.Board = nil
	r.SendAll("stopgame")

	return true
}

// AddPlayer adds the given player to the current game.
func (r *Room) AddPlayer(player *Player) (playerA bool, err error) {
	if r.PlayerA == nil {
		r.PlayerA = player
		playerA = true
	} else if r.PlayerB == nil {
		r.PlayerB = player
		playerA = false
	} else {
		return false, errors.New("room-full")
	}

	return playerA, nil
}

// AddSpectator adds the given player to the game as a spectator.
func (r *Room) AddSpectator(player *Player) {
	r.Spectators = append(r.Spectators, player)
}

// SendAll sends the given type and args to every player and spectator in the
// room.
func (r *Room) SendAll(typ string, args ...string) {
	r.PlayerA.Conn.Send(typ, args...)
	r.PlayerB.Conn.Send(typ, args...)
	r.SendSpectators(typ, args...)
}

// SendSpectators sends the given type and args to every specator in the room.
func (r *Room) SendSpectators(typ string, args ...string) {
	for _, spec := range r.Spectators {
		spec.Conn.Send(typ, args...)
	}
}
