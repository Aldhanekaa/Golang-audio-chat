package server

import (
	"encoding/json"
	"golang-webchat/model"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
)

// AllRooms is the global hashmap for the server
var AllRooms RoomMap

// CreateRoomRequestHandler Create a Room and return roomID

func CreateRoomRequestHandler(w http.ResponseWriter, r *http.Request) {
	// log.Println(r.GetBody())
	log.Println(r.Method)
	if r.Method == "POST" {
		// run
		headerContentType := r.Header.Get("Content-Type")
		if headerContentType != "application/json" {
			json.NewEncoder(w).Encode(struct {
				Message string `json:"message"`
			}{Message: "Content Type is not application/json"})
		}
		return
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")

	roomID := AllRooms.CreateRoom(&model.CreateRoomJSON{
		Id: "",
	})

	type resp struct {
		RoomID string `json:"room_id"`
	}

	json.NewEncoder(w).Encode(resp{RoomID: roomID})
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type broadcastMsg struct {
	Message       map[string]interface{}
	RoomID        string
	Client        *websocket.Conn
	ParticipantId int
}

var broadcast = make(chan broadcastMsg)

func broadcaster() {
	for {
		msg := <-broadcast
		log.Println("MESSAGE ON BROADCASTER: ", msg)
		log.Println("MESSAGE ON BROADCASTER (ROOM ID): ", msg.RoomID)
		log.Println("MESSAGE ON BROADCASTER (ADRESS): ", msg.Client.LocalAddr().String())
		log.Println("MESSAGE ON BROADCASTER (Network): ", msg.Client.LocalAddr().Network())

		for _, client := range AllRooms.Map[msg.RoomID].Participants {
			if client.Conn == msg.Client && msg.Message["action"] == "leave" {
				client.Conn.Close()
				AllRooms.RemoveParticipant(msg.RoomID, msg.ParticipantId)

			}
			// send event to other connected clients in a room
			if client.Conn != msg.Client {
				err := client.Conn.WriteJSON(msg.Message)

				if err != nil {
					log.Println("Broadcast MSG ERROR: ", err)
					log.Println(AllRooms.Map[msg.RoomID])
					client.Conn.Close()
					// return
				}
			}

			// gives participant Id | initial message sent to server
			if client.Conn == msg.Client && msg.Message["ask"] == true {
				client.Conn.WriteJSON(map[string]interface{}{
					"participantId": msg.ParticipantId,
				})
			}
		}
	}
}

// JoinRoomRequestHandler will join the client in a particular room
func JoinRoomRequestHandler(w http.ResponseWriter, r *http.Request) {
	roomID, ok := r.URL.Query()["roomID"]

	if !ok {
		log.Println("roomID missing in URL Parameters")
		return
	}

	if _, ok := AllRooms.Map[roomID[0]]; !ok {
		json.NewEncoder(w).Encode(struct {
			message string `json:"room_id"`
		}{message: "Room not found"})
		return

	}

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Web Socket Upgrade Error", err)
	}

	participantId := AllRooms.InsertIntoRoom(roomID[0], false, ws)

	go broadcaster()

	for {
		var msg broadcastMsg

		err := ws.ReadJSON(&msg.Message)
		if err != nil {
			log.Println("Read Error: ", err)

			if strings.Contains(err.Error(), "websocket: close 1001 (going away)") {
				log.Println("Hey Im an Error!")

				return
			}

			return
		}

		msg.Client = ws
		msg.RoomID = roomID[0]
		msg.ParticipantId = participantId

		log.Println("msg: ", msg)

		broadcast <- msg
	}
}

func GetPeers() {

}
