package main

import (
	"bufio"
	"encoding/json"
	"io"
	"log"

	"golang.org/x/net/websocket"
)

// Connection holds state of a connection with a player.
type Connection struct {
	Chan      chan Message
	ws        *websocket.Conn
	currentID int
}

// Message is a message that is sent between the player and the server.
type Message struct {
	ID        int      `json:"id"`
	Type      string   `json:"type"`
	Arguments []string `json:"args"`
}

func makeConnection(ws *websocket.Conn) *Connection {
	conn := &Connection{
		Chan: make(chan Message),
		ws:   ws,
	}
	reader := bufio.NewReader(ws)

	go func() {
		for {
			raw, err := reader.ReadBytes('\n')
			if err != nil {
				if err != io.EOF {
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

// Send sends a message with the given type and args to the current connection.
func (conn *Connection) Send(typ string, args ...string) (n int, err error) {
	if args == nil {
		args = make([]string, 0)
	}
	conn.currentID++
	return conn.write(Message{
		ID:        conn.currentID,
		Type:      typ,
		Arguments: args,
	})
}

// Reply sends a reply to the given id with the given type and args to the
// current connection.
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
	withlf := append(bytes, '\n')
	return conn.ws.Write(withlf)
}
