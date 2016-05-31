package main

import (
	"bufio"
	"encoding/json"

	"golang.org/x/net/websocket"
)

type Connection struct {
	Chan      chan Message
	ws        *websocket.Conn
	currentID int
}

type Message struct {
	ID        int      `json:"id"`
	Type      string   `json:"type"`
	Arguments []string `json:"args"`
}

func makeConnection(ws *websocket.Conn) *Connection {
	conn := &Connection{
		ws: ws,
	}
	reader := bufio.NewReader(ws)

	go func() {
		for {
			raw, err := reader.ReadBytes('\n')
			if err != nil {
				close(conn.Chan)
				return
			}

			var msg Message
			json.Unmarshal(raw, &msg)
			if msg.ID > conn.currentID {
				conn.currentID = msg.ID
			}
		}
	}()

	return conn
}

func (conn *Connection) Send(typ string, args ...string) (n int, err error) {
	n, err = conn.write(Message{
		ID:        conn.currentID,
		Type:      typ,
		Arguments: args,
	})
	conn.currentID++
	return
}

func (conn *Connection) Reply(id int, typ string, args ...string) (n int, err error) {
	return conn.write(Message{
		ID:        id,
		Type:      typ,
		Arguments: args,
	})
}

func (conn *Connection) write(msg Message) (n int, err error) {
	bytes, err := json.Marshal(msg)
	if err != nil {
		return -1, err
	}
	return conn.ws.Write(bytes)
}
