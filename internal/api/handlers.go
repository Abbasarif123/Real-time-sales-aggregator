package api

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"Real-timesales/internal/models"
	"Real-timesales/internal/state"
	"Real-timesales/internal/streaming"
)

// HandleIngest processes incoming synthetic sales data from the producer
func HandleIngest(buffer *state.RingBuffer, hub *streaming.Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		//restrict this endpoint to POST requests only
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		//stream decode the raw JSON payload directly from the request body into a Transaction struct
		var tx models.Transaction
		if err := json.NewDecoder(r.Body).Decode(&tx); err != nil {
			//if the JSON is malformed or types mismatched then return a error
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		//close the request body stream to release underlying network resources
		defer r.Body.Close()

		//if the incoming transaction's timestamp is zero, set it to the current time
		if tx.Timestamp.IsZero() {
			tx.Timestamp = time.Now()
		}

		//add the transaction to the ring buffer
		buffer.Add(tx)

		//calculate a fresh rolling snapshot
		metrics := buffer.GetMetrics()

		//push rhe newly calculated snapshot to the hub for broadcasting to all connected clients
		select {
		//if hubs broadcast has buffer space, send the metrics
		case hub.Broadcast <- metrics:

		//if its full then instead of blocking this HTTP handler, drop the frame
		default:
			log.Println("Hub broadcast channel is full, dropping snapshot frame")
		}

		//return HTTP 200 OK with a JSON confirmation to the data producer
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}
}
