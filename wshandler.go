package main

import "golang.org/x/net/websocket"

func WsHandler(ws *websocket.Conn) {
	player := makePlayer(ws)
	var currentRoom *Room
	waitch := make(chan int)

	go func() {
		for {
			msg, open := <-player.Conn.Chan
			if !open {
				close(waitch)
				return
			}

			reply := func(typ string, args ...string) {
				player.Conn.Reply(msg.ID, typ, args...)
			}

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

				other.Conn.Send(typ, args...)
				for _, spectator := range currentRoom.Spectators {
					spectator.Conn.Send(typ, args...)
				}
			}

			switch msg.Type {
			case "ping":
				reply("pong")
			case "setnick":
				player.Nickname = msg.Arguments[0]
				notifyOthers()
				reply("ok", player.Nickname)

			case "joinroom":
				room := getRoom(msg.Arguments[0])
				if room == nil {
					reply("error", "not-found")
				} else if room.PlayerA == nil {
					room.PlayerA = player
					currentRoom = room
					reply("ok")
				} else if room.PlayerB == nil {
					room.PlayerB = player
					currentRoom = room
					reply("ok")
				} else {
					reply("error", "room-full")
				}
			case "leaveroom":
				if currentRoom != nil {
					if currentRoom.PlayerA == player {
						currentRoom.PlayerA = currentRoom.PlayerB
					}
					currentRoom.PlayerB = nil
					currentRoom = nil
					reply("ok")
				} else {
					reply("error", "not-in-room")
				}
			}
		}
	}()

	<-waitch
}
