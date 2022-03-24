package server

import (
	"log"
	"math/rand"
	"strings"
	"sync"
	"time"

	model "golang-webchat/model"

	"github.com/gorilla/websocket"
)

var IdOptions = []string{"d"}

// Participant describes a single entity in the hashmap
type Participant struct {
	Host bool

	Conn *websocket.Conn
}

// rId (Room Id) | pId (Participant Id)
func (p *Participant) removeParticipant(rId string, pId string, roomMap *RoomMap) {

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
	r.Map[roomID].Participants[1] = Participant{
		Host: true,
	}

	log.Println(r.Map[roomID])
	return roomID
}

// InsertIntoRoom will create a participant and add it in the hashmap
func (r *RoomMap) InsertIntoRoom(roomID string, host bool, conn *websocket.Conn) {
	r.Mutex.Lock()
	defer r.Mutex.Unlock()

	participantId := r.Map[roomID].LastUserId + 1

	// make instance of Participant
	p := Participant{host, conn}

	if entry, ok := r.Map[roomID]; ok {
		entry.LastUserId += 1
		r.Map[roomID] = entry
	}

	log.Println(r.Map[roomID])

	r.Map[roomID].Participants[participantId] = p
	log.Println(r.Map[roomID])

}

// DeleteRoom deletes the room with the roomID
func (r *RoomMap) DeleteRoom(roomID string) {
	r.Mutex.Lock()
	defer r.Mutex.Unlock()

	delete(r.Map, roomID)
}
