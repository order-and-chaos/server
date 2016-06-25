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
		}

		return errors.New("not-in-room")
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

			handleCommand := func(typ string, argCount int, requiresGame bool, fn func()) {
				if msg.Type != typ {
					return
				}

				if len(msg.Arguments) != argCount {
					reply("error", "format")
					return
				}

				if requiresGame {
					if currentRoom == nil {
						reply("error", "not-in-room")
						return
					} else if currentRoom.Board == nil {
						reply("error", "not-in-game")
						return
					}
				}

				fn()
			}

			handleCommand("ping", 0, false, func() {
				reply("pong")
			})

			handleCommand("setnick", 1, false, func() {
				oldNick := player.Nickname
				player.Nickname = msg.Arguments[0]
				notifyOthers("setnick", oldNick, player.Nickname)
				reply("ok", player.Nickname)

			})
			handleCommand("getnick", 0, false, func() {
				reply("ok", player.Nickname)
			})

			handleCommand("joinroom", 1, false, func() {
				_, err := joinRoom(msg.Arguments[0])
				if err != nil {
					reply("error", err.Error())
				} else {
					reply("ok")
				}
			})
			handleCommand("leaveroom", 0, false, func() {
				err := leaveRoom()
				if err != nil {
					reply("error", err.Error())
				} else {
					reply("ok")
				}
			})
			handleCommand("makeroom", 0, false, func() {
				room := mkRoom()
				joinRoom(room.ID)
				reply("ok", room.ID)
			})

			handleCommand("startgame", 0, false, func() {
				if currentRoom == nil {
					reply("error", "not-in-room")
				} else if currentRoom.Board != nil {
					reply("error", "already-in-game")
				} else if !currentRoom.StartGame() {
					reply("error", "not-ready")
				}
			})
			handleCommand("stopgame", 0, true, func() {
				currentRoom.StopGame()
			})

			handleCommand("getboard", 0, true, func() {
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
			})

			handleCommand("getonturn", 0, true, func() {
				if currentRoom.Board.Onturn == Order {
					reply("ok", "order")
				} else {
					reply("ok", "chaos")
				}
			})

			handleCommand("applymove", 2, true, func() {
				var stone Cell
				if msg.Arguments[0] == "O" {
					stone = OO
				} else if msg.Arguments[0] == "X" {
					stone = XX
				} else {
					reply("error", "format-error")
					return
				}

				if (player == currentRoom.PlayerA &&
					currentRoom.Board.Onturn != currentRoom.RoleA) ||
					(player != currentRoom.PlayerA &&
						currentRoom.Board.Onturn == currentRoom.RoleA) {
					reply("error", "not-on-turn")
					return
				}

				pos, err := strconv.Atoi(msg.Arguments[1])
				if err != nil || pos < 0 || pos >= N*N {
					reply("error", "format-error")
					return
				}
				empty, _ := currentRoom.Board.IsEmpty(pos)
				if empty {
					reply("error", "cell-not-empty")
					return
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
			})
		}
	}()

	<-waitch
}
