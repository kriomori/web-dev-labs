package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"lab1/internal/model"
)

// FetchPosts скачивает данные с API и декодирует JSON
func FetchPosts() ([]model.Post, error) {
	resp, err := http.Get("https://jsonplaceholder.typicode.com/posts")
	if err != nil {
		return nil, fmt.Errorf("ошибка HTTP запроса: %w", err)
	}
	defer resp.Body.Close()

	var posts []model.Post
	if err := json.NewDecoder(resp.Body).Decode(&posts); err != nil {
		return nil, fmt.Errorf("ошибка парсинга JSON: %w", err)
	}

	return posts, nil
}
