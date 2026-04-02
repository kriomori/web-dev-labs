package main

import (
	"bufio"
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/gorilla/websocket"
)

type Message struct {
	Type     string `json:"type"`
	Token    string `json:"token,omitempty"`
	Username string `json:"username,omitempty"`
	Text     string `json:"text,omitempty"`
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Использование: go run client.go <username> <room>")
		return
	}
	username := os.Args[1]
	room := os.Args[2]

	u := url.URL{
		Scheme:   "ws",
		Host:     "localhost:8088",
		Path:     "/ws",
		RawQuery: fmt.Sprintf("room=%s&username=%s", room, username),
	}

	fmt.Printf("Подключение к %s...\n", u.String())

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("Ошибка подключения:", err)
	}
	defer conn.Close()

	go func() {
		for {
			var msg Message
			err := conn.ReadJSON(&msg)
			if err != nil {
				log.Println("\nОтключено от сервера.")
				os.Exit(0)
			}

			switch msg.Type {
			case "auth":
				fmt.Printf("[Система] %s (Ваш auth-токен: %s)\n", msg.Text, msg.Token)
			case "system":
				fmt.Printf("\n[Система] %s\n> ", msg.Text)
			case "message":
				fmt.Printf("\n[%s]: %s\n> ", msg.Username, msg.Text)
			}
		}
	}()

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("> ")
	for scanner.Scan() {
		text := scanner.Text()
		if text == "" {
			fmt.Print("> ")
			continue
		}

		msg := Message{
			Type: "message",
			Text: text,
		}

		err := conn.WriteJSON(msg)
		if err != nil {
			log.Println("Ошибка отправки:", err)
			return
		}
		fmt.Print("> ")
	}
}
