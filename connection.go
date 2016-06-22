package main

import (
	"bufio"
	"encoding/json"
	"log"

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
		Chan: make(chan Message),
		ws: ws,
	}
	reader := bufio.NewReader(ws)

	go func() {
		for {
			raw, err := reader.ReadBytes('\n')
			if err != nil {
				if err.Error() != "EOF" {
					log.Printf("error while reading from connection: %#v\n", err)
				}
				close(conn.Chan)
				return
			}

			var msg Message
			if json.Unmarshal(raw, &msg) != nil {
				log.Printf("invalid message received: %s\n", raw)
				continue
			}
			if msg.ID > conn.currentID {
				conn.currentID = msg.ID
			}
			conn.Chan <- msg
		}
	}()

	return conn
}

func (conn *Connection) Send(typ string, args ...string) (n int, err error) {
	if args == nil {
		args = make([]string, 0)
	}
	conn.currentID++
	n, err = conn.write(Message{
		ID:        conn.currentID,
		Type:      typ,
		Arguments: args,
	})
	return
}

func (conn *Connection) Reply(id int, typ string, args ...string) (n int, err error) {
	if args == nil {
		args = make([]string, 0)
	}
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
