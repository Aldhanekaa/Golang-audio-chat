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

func broadcaster(broadcast *chan broadcastMsg) {
	for {
		// var removeParticipant bool

		msg := <-*broadcast
		// log.Println("MESSAGE ON BROADCASTER: ", msg)
		log.Println("MESSAGE ON BROADCASTER (ROOM ID): ", msg.RoomID)
		log.Println("MESSAGE ON BROADCASTER (ADRESS): ", msg.Client.LocalAddr().String())
		log.Println("MESSAGE ON BROADCASTER (Network): ", msg.Client.LocalAddr().Network())

		for _, client := range AllRooms.Map[msg.RoomID].Participants {

			log.Println("ASK!")

			// gives participant Id | initial message sent to server
			if client.Conn == msg.Client && msg.Message["ask"] == true {
				client.Conn.WriteJSON(map[string]interface{}{
					"participantId": msg.ParticipantId,
				})

			}
			log.Println("OTHER USER")

			// send event to other connected clients in a room
			if client.Conn != msg.Client {
				// log.Println(msg.Message)
				err := client.Conn.WriteJSON(msg.Message)

				if err != nil {
					log.Println("Broadcast MSG ERROR: ", err)
					log.Println(AllRooms.Map[msg.RoomID])
					client.Conn.Close()
					return
				}
			}
			log.Println("LEFT CHAT")
			// log.Println(client, msg)

			if client.Conn == msg.Client && msg.Message["action"] == "leave" {
				// log.Println("wait")

				err := client.Conn.Close()

				if err != nil {
					log.Println("Error Closing WS", err.Error())
				} else {
					AllRooms.RemoveParticipant(msg.RoomID, msg.ParticipantId)

				}
				// fmt.Println(AllRooms.Map)

				break
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

	roomId := roomID[0]

	participantId := AllRooms.InsertIntoRoom(roomID[0], false, ws)
	var broadcast = make(chan broadcastMsg)

	go broadcaster(&broadcast)

	for {
		log.Println("Participant with Id: ", participantId, " Broadcast a message")

		var msg broadcastMsg

		err := ws.ReadJSON(&msg.Message)
		if err != nil {
			log.Println("Read Error: ", err)

			if strings.Contains(err.Error(), "websocket: close 1001") || strings.Contains(err.Error(), "websocket: close 1005") || strings.Contains(err.Error(), "websocket: close 1006") {
				log.Println("ERROR TAU ", err)
				// log.Println("Hey Im an Error!")
				RemoveParticipant(roomId, participantId, &AllRooms)
				ws.Close()

			}

			return
		}

		msg.Client = ws
		msg.RoomID = roomID[0]
		msg.ParticipantId = participantId

		// log.Println("msg: ", msg)

		broadcast <- msg
	}
}

func GetPeers() {

}
