package main

import (
	"log"
	"net/http"

	"Real-timesales/internal/api"
	"Real-timesales/internal/state"
	"Real-timesales/internal/streaming"
)

func main() {
	//initialse dependencies

	buffer := state.NewRingBuffer(1000) //buffer size of 1000 transactions
	hub := streaming.NewHub()           //central hub for managing websocket clients and broadcasting messages

	//start the websocket manager
	go hub.Run()

	//setup the router
	mux := http.NewServeMux()

	//endpoint A ie the websocket connection for the react dashboard
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		streaming.ServeWs(hub, w, r)
	})

	//endpoint B: the POST endpoint passing in the buffer ad hub dependencies
	mux.HandleFunc("/ingest", api.HandleIngest(buffer, hub))

	//start the server
	port := ":8080"
	log.Printf("Server starting on http://localhost%s", port)
	if err := http.ListenAndServe(port, mux); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
