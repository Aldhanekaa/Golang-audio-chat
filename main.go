package main

import (
	"encoding/json"
	"fmt"
	"golang-webchat/server"
	"log"
	"net/http"
)

func main() {
	server.AllRooms.Init()

	http.HandleFunc("/create", server.CreateRoomRequestHandler)
	http.HandleFunc("/join", server.JoinRoomRequestHandler)
	http.HandleFunc("/checkroom", func(w http.ResponseWriter, r *http.Request) {
		roomID, ok := r.URL.Query()["roomID"]
		if !ok {
			log.Println("roomID is missing when want to check room")
			json.NewEncoder(w).Encode(struct {
				Message string `json:"message"`
			}{Message: "roomID is missing when want to check room"})
			return
		}

		if _, ok := server.AllRooms.Map[roomID[0]]; !ok {
			fmt.Fprint(w, "not found")

			return
		}

		fmt.Fprint(w, "found")
		return

	})

	log.Println("Starting Server on Port 8000")
	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatal("FATAL ERROR : ", err)
	}
}
