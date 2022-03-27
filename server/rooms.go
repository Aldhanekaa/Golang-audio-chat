package server

import (
	"log"
	"math/rand"
	"reflect"
	"strings"
	"sync"
	"time"

	model "golang-webchat/model"

	"github.com/gorilla/websocket"
)

// Participant describes a single entity in the hashmap
type Participant struct {
	Host bool

	Conn *websocket.Conn
}

// room | Participants in the room [participant id] -> Participant
type Room struct {
	LastUserId   int
	Participants map[int]Participant
}

// RoomMap is the main hashmap [roomID string] -> [[]Room]
type RoomMap struct {
	Mutex sync.RWMutex
	Map   map[string]Room
}

// Init initialises the RoomMap struct
func (r *RoomMap) Init() {
	r.Map = make(map[string]Room)
}

func (r *RoomMap) InitParticipations(roomId string) {

	if room, ok := r.Map[roomId]; ok {
		room.Participants = make(map[int]Participant)
		r.Map[roomId] = room
	}

}

var RemoveParticipant = func(roomId string, participantId int, r interface{}) {
	typeofr := reflect.TypeOf(r).String()
	log.Println(r)
	var roomMap *RoomMap

	if typeofr == "*server.RoomMap" {
		roomMap = r.(*RoomMap)
	} else {
		roomMap = &AllRooms
	}
	// log.Println("tee!", roomMap.Map)

	var delRoom bool = false
	// var rooom Room

	// rooom = Room{}

	// rooom = roomMap.Map[roomId]
	// log.Println("yooo!", roomId, roomMap.Map[roomId])

	room, ok := roomMap.Map[roomId]
	if ok {
		log.Println("REMOVE PARTICIPANT!!")
		log.Println(roomMap.Map[roomId])

		delete(room.Participants, participantId)
		log.Println(roomMap.Map[roomId])

		if len(room.Participants) == 0 {
			delRoom = true
		}

		roomMap.Map[roomId] = room
	}

	if delRoom {
		delete(roomMap.Map, roomId)

	}

}

// rId (Room Id) | pId (Participant Id)
func (r *RoomMap) RemoveParticipant(roomId string, participantId int) {

	RemoveParticipant(roomId, participantId, r)

}

// Get will return the array of participants in the room
func (r *RoomMap) GetParticipants(roomID string) map[int]Participant {
	r.Mutex.RLock()
	defer r.Mutex.RUnlock()

	return r.Map[roomID].Participants
}

// CreateRoom generate a unique room ID and return it -> insert it in the hashmap
func (r *RoomMap) CreateRoom(randomise *model.CreateRoomJSON) string {
	r.Mutex.Lock()
	defer r.Mutex.Unlock()

	randomise.Id = strings.Trim(randomise.Id, " ")

	if len(randomise.Id) < 5 {

	}

	rand.Seed(time.Now().UnixNano())

	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")
	b := make([]rune, 8)

	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}

	roomID := string(b)
	r.Map[roomID] = Room{LastUserId: 0}
	r.InitParticipations(roomID)
	return roomID
}

// InsertIntoRoom will create a participant and add it in the hashmap
func (r *RoomMap) InsertIntoRoom(roomID string, host bool, conn *websocket.Conn) int {
	r.Mutex.Lock()
	defer r.Mutex.Unlock()

	participantId := r.Map[roomID].LastUserId + 1

	// make instance of Participant
	p := Participant{host, conn}

	if entry, ok := r.Map[roomID]; ok {
		entry.LastUserId += 1
		r.Map[roomID] = entry
	}

	r.Map[roomID].Participants[participantId] = p
	log.Println(r.Map[roomID])

	return participantId
}

// DeleteRoom deletes the room with the roomID
func (r *RoomMap) DeleteRoom(roomID string) {
	r.Mutex.Lock()
	defer r.Mutex.Unlock()

	delete(r.Map, roomID)
}
