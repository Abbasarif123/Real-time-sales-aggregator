package streaming

import "Real-timesales/internal/models"

// Hub maintains the set of active clients and broadcasts messages to them
type Hub struct {
	Clients    map[*Client]bool        //registered clients
	Broadcast  chan models.KPISnapshot //inbound channel from our data ingestion endpoint
	Register   chan *Client
	Unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		Broadcast:  make(chan models.KPISnapshot),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Clients:    make(map[*Client]bool),
	}
}

func (h *Hub) Run() {
	for { //infinite loop to continuously process incoming lifecycle events and messages
		select { //pause until a case is fulfilled
		case client := <-h.Register: //client registers
			h.Clients[client] = true

		case client := <-h.Unregister: //client unregisters
			if _, ok := h.Clients[client]; ok { //verify if it exists in the registry then delete it
				delete(h.Clients, client)
				close(client.Send)
			}

		case message := <-h.Broadcast: //broadcasting ie when a new payload is ready to send to all connected clients
			for client := range h.Clients { //iterate through every registered client
				select { //attempt to push the message into the clients buffered send channel
				case client.Send <- message:

				// safety: if client.Send's buffer is full, the default case executes immediately instead of blocking the entire Hub
				default:
					close(client.Send)
					delete(h.Clients, client)
				}
			}
		}
	}
}
