package main

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type Message struct {
	Type     string `json:"type"`
	Token    string `json:"token,omitempty"`
	Username string `json:"username,omitempty"`
	Text     string `json:"text,omitempty"`
}

type Client struct {
	conn     *websocket.Conn
	username string
	token    string
	send     chan Message
}

type Room struct {
	name       string
	clients    map[*Client]bool
	broadcast  chan Message
	register   chan *Client
	unregister chan *Client
}

var (
	rooms   = make(map[string]*Room)
	roomsMu sync.Mutex
)

func getRoom(name string) *Room {
	roomsMu.Lock()
	defer roomsMu.Unlock()

	if room, ok := rooms[name]; ok {
		return room
	}

	room := &Room{
		name:       name,
		clients:    make(map[*Client]bool),
		broadcast:  make(chan Message),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
	rooms[name] = room

	go room.run()
	return room
}

func (r *Room) run() {
	for {
		select {
		case client := <-r.register:
			r.clients[client] = true
		case client := <-r.unregister:
			if _, ok := r.clients[client]; ok {
				delete(r.clients, client)
				close(client.send)
			}
		case message := <-r.broadcast:
			for client := range r.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(r.clients, client)
				}
			}
		}
	}
}

func generateToken() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func serveWs(w http.ResponseWriter, r *http.Request) {
	roomName := r.URL.Query().Get("room")
	username := r.URL.Query().Get("username")

	if roomName == "" || username == "" {
		http.Error(w, "Параметры room и username обязательны", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Ошибка апгрейда соединения:", err)
		return
	}

	room := getRoom(roomName)
	token := generateToken()

	client := &Client{
		conn:     conn,
		username: username,
		token:    token,
		send:     make(chan Message, 256),
	}

	room.register <- client

	client.send <- Message{Type: "auth", Token: token, Text: "Вы подключились к комнате " + roomName}
	room.broadcast <- Message{Type: "system", Text: username + " присоединился к чату"}

	go client.write()
	go client.read(room)
}

func (c *Client) read(room *Room) {
	defer func() {
		room.unregister <- c
		c.conn.Close()
		room.broadcast <- Message{Type: "system", Text: c.username + " покинул чат"}
	}()

	for {
		var msg Message
		err := c.conn.ReadJSON(&msg)
		if err != nil {
			break
		}
		msg.Type = "message"
		msg.Username = c.username
		room.broadcast <- msg
	}
}

func (c *Client) write() {
	defer c.conn.Close()
	for msg := range c.send {
		err := c.conn.WriteJSON(msg)
		if err != nil {
			return
		}
	}
}

func main() {
	http.HandleFunc("/ws", serveWs)
	log.Println("Сервер запущен на :8088")
	if err := http.ListenAndServe(":8088", nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
