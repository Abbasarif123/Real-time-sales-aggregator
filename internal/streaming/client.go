package streaming

import (
	"log"
	"net/http"
	"time"

	"Real-timesales/internal/models"

	"github.com/gorilla/websocket"
)

// upgrader configures HTTP-to-WebSocket protocol switching parameters
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024, // 1 KB buffer for inbound frame chunking
	WriteBufferSize: 1024, // 1 KB buffer for outbound frame chunking

	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// client acts as the per-connection state bridge between the central Hub and the physical network socket
type Client struct {
	hub  *Hub                    // reference to the central coordinator
	conn *websocket.Conn         // active underlying WebSocket connection
	Send chan models.KPISnapshot // buffered outbound message queue (prevents blocking during Hub broadcasts)
}

// writePump pumps messages from the hub to the websocket connectiom
func (c *Client) writePump() {
	defer func() { //ensure network socket is closed during the writing phase
		c.conn.Close()
	}()

	for {
		select {
		//recieve snapshot payloads queued up by the hub
		case message, ok := <-c.Send:
			//set a 10 second deadlines, if the network write stalls longer than the limit then the connection aborts
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				//hub closed the channel
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			// write the KPISnapshot as JSON to the frontend
			err := c.conn.WriteJSON(message)
			if err != nil {
				log.Printf("Error writing JSON to websocket: %v", err)
				return //terminate the routine on write failure
			}
		}
	}
}

// ServeWs is the HTTP route handler that upgrades regular HTTP requests into full duplex websocket connections
func ServeWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	//perform websocket handshake
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	//define client with a 256 item buffered outbound channel to absorb the bursts in traffic
	client := &Client{
		hub:  hub,
		conn: conn,
		Send: make(chan models.KPISnapshot, 256),
	}
	client.hub.Register <- client

	// start the write pump in a new goroutine so it runs concurrently
	go client.writePump()

	//no expectation that the frontend is gonna send us messages
	// but we must read from the connection to process close events gracefully
	go func() {
		defer func() {
			client.hub.Unregister <- client
			client.conn.Close()
		}()
		for {
			_, _, err := client.conn.ReadMessage()
			if err != nil {
				break // disconnected
			}
		}
	}()
}
