package veto

import (
	"log"

	"github.com/gorilla/websocket"
)

const sendBufferSize = 16

type Client struct {
	conn *websocket.Conn
	send chan []byte
	done chan struct{}
}

func newClient(conn *websocket.Conn) *Client {
	return &Client{
		conn: conn,
		send: make(chan []byte, sendBufferSize),
		done: make(chan struct{}),
	}
}

func (c *Client) Send(payload []byte) {
	select {
	case c.send <- payload:
	default:
		c.Close()
	}
}

func (c *Client) Close() {
	select {
	case <-c.done:
	default:
		close(c.done)
	}
}

func (c *Client) writePump() {
	for {
		select {
		case payload := <-c.send:
			if err := c.conn.WriteMessage(websocket.TextMessage, payload); err != nil {
				c.Close()
				return
			}
		case <-c.done:
			if err := c.conn.Close(); err != nil {
				log.Printf("failed to close websocket connection: %v", err)
			}
			return
		}
	}
}
