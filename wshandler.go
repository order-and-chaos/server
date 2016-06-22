package main

import (
	"errors"

	"golang.org/x/net/websocket"
)

func WsHandler(ws *websocket.Conn) {
	player := makePlayer(ws)
	var currentRoom *Room
	waitch := make(chan int)

	notifyOthers := func(typ string, args ...string) {
		if currentRoom == nil {
			return
		}

		var other *Player
		if currentRoom.PlayerA == player {
			other = currentRoom.PlayerB
		} else {
			other = currentRoom.PlayerA
		}

		if other != nil {
			other.Conn.Send(typ, args...)
		}

		for _, spectator := range currentRoom.Spectators {
			spectator.Conn.Send(typ, args...)
		}
	}

	leaveRoom := func() error {
		if currentRoom != nil {
			if currentRoom.PlayerA == player {
				currentRoom.PlayerA = currentRoom.PlayerB
			}
			currentRoom.PlayerB = nil
			currentRoom = nil
			notifyOthers("joinroom", player.Nickname)
			return nil
		} else {
			return errors.New("not-in-room")
		}
	}

	joinRoom := func(id string) (*Room, error) {
		if currentRoom != nil {
			leaveRoom()
		}
		room := getRoom(id)
		if room == nil {
			return nil, errors.New("not-found")
		} else if room.PlayerA == nil {
			room.PlayerA = player
			currentRoom = room
		} else if room.PlayerB == nil {
			room.PlayerB = player
			currentRoom = room
		} else {
			return nil, errors.New("room-full")
		}
		return room, nil
	}

	go func() {
		for {
			msg, open := <-player.Conn.Chan
			if !open {
				leaveRoom()
				close(waitch)
				return
			}

			reply := func(typ string, args ...string) {
				player.Conn.Reply(msg.ID, typ, args...)
			}

			argreply := func(count int) bool {
				if len(msg.Arguments) != count {
					reply("error", "format")
					return false
				}
				return true
			}

			switch msg.Type {
			case "ping":
				if !argreply(0) {continue}
				reply("pong")

			case "setnick":
				if !argreply(1) {continue}
				oldNick := player.Nickname
				player.Nickname = msg.Arguments[0]
				notifyOthers("setnick", oldNick, player.Nickname)
				reply("ok", player.Nickname)

			case "getnick":
				if !argreply(0) {continue}
				reply("ok", player.Nickname)

			case "joinroom":
				if !argreply(1) {continue}
				_, err := joinRoom(msg.Arguments[0])
				if err != nil {
					reply("error", err.Error())
				} else {
					reply("ok")
				}

			case "leaveroom":
				if !argreply(0) {continue}
				err := leaveRoom()
				if err != nil {
					reply("error", err.Error())
				} else {
					reply("ok")
				}

			case "makeroom":
				if !argreply(0) {continue}
				room := mkRoom()
				joinRoom(room.ID)
				reply("ok", room.ID)
			}
		}
	}()

	<-waitch
}
