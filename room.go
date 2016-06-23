package main

type Room struct {
	ID               string
	PlayerA, PlayerB *Player
	Spectators       []*Player
	OnMove           int
	Started          bool
}

// Start starts the game of this room.
func (r *Room) Start() bool {
	if r.Started {
		return true
	}

	if r.PlayerA == nil || r.PlayerB == nil {
		return false
	}

	r.OnMove = 0
	r.sendAll("gamestart")

	return true
}

func (r *Room) sendAll(typ string, args ...string) {
	r.PlayerA.Conn.Send(typ, args...)
	r.PlayerB.Conn.Send(typ, args...)
	r.sendSpectators(typ, args...)
}

func (r *Room) sendSpectators(typ string, args ...string) {
	for _, spec := range r.Spectators {
		spec.Conn.Send(typ, args...)
	}
}
