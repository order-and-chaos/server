package main

import (
	"errors"
	"strconv"

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

		currentRoom.SendSpectators(typ, args...)
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

			argcheck := func(count int) bool {
				if len(msg.Arguments) != count {
					reply("error", "format")
					return false
				}
				return true
			}

			gamecheck := func() bool {
				if currentRoom == nil {
					reply("error", "not-in-room")
				} else if currentRoom.Board == nil {
					reply("error", "not-in-game")
				} else {
					return true
				}
				return false
			}

			switch msg.Type {
			case "ping":
				if !argcheck(0) {continue}
				reply("pong")

			case "setnick":
				if !argcheck(1) {continue}
				oldNick := player.Nickname
				player.Nickname = msg.Arguments[0]
				notifyOthers("setnick", oldNick, player.Nickname)
				reply("ok", player.Nickname)

			case "getnick":
				if !argcheck(0) {continue}
				reply("ok", player.Nickname)

			case "joinroom":
				if !argcheck(1) {continue}
				_, err := joinRoom(msg.Arguments[0])
				if err != nil {
					reply("error", err.Error())
				} else {
					reply("ok")
				}

			case "leaveroom":
				if !argcheck(0) {continue}
				err := leaveRoom()
				if err != nil {
					reply("error", err.Error())
				} else {
					reply("ok")
				}

			case "makeroom":
				if !argcheck(0) {continue}
				room := mkRoom()
				joinRoom(room.ID)
				reply("ok", room.ID)

			case "startgame":
				if !argcheck(0) {continue}
				if currentRoom == nil {
					reply("error", "not-in-room")
				} else if currentRoom.Board != nil {
					reply("error", "already-in-game")
				} else if !currentRoom.StartGame() {
					reply("error", "not-ready")
				}

			case "stopgame":
				if !argcheck(0) || !gamecheck() {continue}
				currentRoom.StopGame()

			case "getboard":
				if !argcheck(0) || !gamecheck() {continue}
				acc := ""
				for i := 0; i < N*N; i++ {
					if currentRoom.Board.Cells[i] == OO {
						acc += "O"
					} else if currentRoom.Board.Cells[i] == XX {
						acc += "X"
					} else {
						acc += " "
					}
				}
				reply("ok", acc)

			case "getonturn":
				if !argcheck(0) || !gamecheck() {continue}
				if currentRoom.Board.Onturn == Order {
					reply("ok", "order")
				} else {
					reply("ok", "chaos")
				}

			case "applymove":
				if !argcheck(2) || !gamecheck() {continue}

				var stone Cell
				if msg.Arguments[0] == "O" {
					stone = OO
				} else if msg.Arguments[0] == "X" {
					stone = XX
				} else {
					reply("error", "format-error")
					continue
				}

				if (player == currentRoom.PlayerA &&
				    currentRoom.Board.Onturn != currentRoom.RoleA) ||
				   (player != currentRoom.PlayerA &&
				    currentRoom.Board.Onturn == currentRoom.RoleA) {
					reply("error", "not-on-turn")
					continue
				}

				pos, err := strconv.Atoi(msg.Arguments[1])
				if err != nil || pos < 0 || pos >= N*N {
					reply("error", "format-error")
					continue
				}
				empty, _ := currentRoom.Board.IsEmpty(pos)
				if empty {
					reply("error", "cell-not-empty")
					continue
				}

				currentRoom.Board.ApplyMove(stone, pos)
				reply("ok")

				role, win := currentRoom.Board.CheckWin()
				if win {
					if role == Order {
						currentRoom.SendAll("win", "order")
					} else {
						currentRoom.SendAll("win", "chaos")
					}
					currentRoom.StopGame()
				}
			}
		}
	}()

	<-waitch
}
