package main

import (
	"errors"
	"strconv"

	"golang.org/x/net/websocket"
)

// WsHandler handles the ws connection from the http server.
func WsHandler(ws *websocket.Conn) {
	player := makePlayer(ws)
	var currentRoom *Room
	waitch := make(chan int)

	// TODO: handle ping/pong timeouts

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

			notifyOthers("leaveroom", player.Nickname)

			return nil
		}

		return errors.New("not-in-room")
	}

	joinRoom := func(id string) (playerA bool, err error) {
		if currentRoom != nil {
			leaveRoom()
		}

		room := getRoom(id)
		if room == nil {
			return false, errors.New("not-found")
		}

		playerA, err = room.AddPlayer(player)
		if err != nil {
			return false, err
		}

		currentRoom = room

		var str string
		if playerA {
			str = "0"
		} else {
			str = "1"
		}
		notifyOthers("joinroom", player.Nickname, str)

		return playerA, nil
	}

	go func() {
		for {
			msg, open := <-player.Conn.Chan
			if !open {
				close(waitch)
				return
			}
			handled := false

			reply := func(typ string, args ...string) {
				player.Conn.Reply(msg.ID, typ, args...)
			}

			handleCommand := func(typ string, argCount int, fn func()) {
				if msg.Type != typ {
					return
				}
				handled = true

				if len(msg.Arguments) != argCount {
					reply("error", "format-nargs")
					return
				}

				fn()
			}

			handleRoomCommand := func(typ string, argCount int, fn func()) {
				handleCommand(typ, argCount, func() {
					if currentRoom == nil {
						reply("error", "not-in-room")
						return
					}

					fn()
				})
			}

			handleGameCommand := func(typ string, argCount int, fn func()) {
				handleRoomCommand(typ, argCount, func() {
					if currentRoom.Board == nil {
						reply("error", "not-in-game")
						return
					}

					fn()
				})
			}

			handleCommand("ping", 0, func() { //HC [] pong []
				reply("pong")
			})

			handleCommand("setnick", 1, func() { //HC [nick] ok [nick] error [err]
				old := player.Nickname
				new := msg.Arguments[0]
				if old == new {
					reply("ok", old)
					return
				}

				if nicknameExists(new) {
					reply("error", "nick-taken")
					return
				}

				player.Nickname = new
				notifyOthers("setnick", old, new)
				reply("ok", new)
			})
			handleCommand("getnick", 0, func() { //HC [] ok [nick]
				reply("ok", player.Nickname)
			})

			handleCommand("joinroom", 1, func() { //HC [id] ok [othernick?] error [err]
				playerA, err := joinRoom(msg.Arguments[0])
				if err != nil {
					reply("error", err.Error())
					return
				}

				if playerA {
					reply("ok")
				} else {
					reply("ok", currentRoom.PlayerA.Nickname)
				}
			})
			handleCommand("spectateroom", 1, func() { //HC [id] ok [] error [err]
				if currentRoom != nil {
					leaveRoom()
				}

				room := getRoom(msg.Arguments[0])
				if room == nil {
					reply("error", "not-found")
					return
				}

				room.AddSpectator(player)
				currentRoom = room
				reply("ok")
			})
			handleCommand("leaveroom", 0, func() { //HC [] ok [] error [err]
				err := leaveRoom()
				if err != nil {
					reply("error", err.Error())
				} else {
					reply("ok")
				}
			})
			handleCommand("makeroom", 0, func() { //HC [] ok []
				room := makeRoom()
				joinRoom(room.ID)
				reply("ok", room.ID)
			})

			handleRoomCommand("sendroomchat", 1, func() { //HC [msg] ok []
				notifyOthers("chatmessage", player.Nickname, currentRoom.ID, msg.Arguments[0])
				reply("ok")
			})

			handleRoomCommand("startgame", 0, func() { //HC [] ok [] error [err]
				if currentRoom.Board != nil {
					reply("error", "already-in-game")
				} else if !currentRoom.StartGame() {
					reply("error", "not-ready")
				}
				reply("ok")
			})
			handleGameCommand("stopgame", 0, func() { //HC [] ok []
				currentRoom.StopGame()
				reply("ok")
			})

			handleGameCommand("getboard", 0, func() { //HC [] ok [boardrepr]
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

			handleGameCommand("getonturn", 0, func() { //HC [] ok [order/chaos,1/2]
				var num string
				if currentRoom.RoleA == currentRoom.Board.Onturn {
					num = "1"
				} else {
					num = "2"
				}
				if currentRoom.Board.Onturn == Order {
					reply("ok", "order", num)
				} else {
					reply("ok", "chaos", num)
				}
			})

			handleGameCommand("applymove", 2, func() { //HC [O/X,pos] ok [] error [err]
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

			if !handled {
				reply("error", "unknown-command", msg.Type)
			}
		}
	}()

	<-waitch

	leaveRoom()

	// remove player from players slice
	var index int
	for i, p := range players {
		if p == player {
			index = i
		}
	}
	players = append(players[:index], players[index+1:]...)
}
