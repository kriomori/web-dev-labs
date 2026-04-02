package pipeline

import (
	"strings"
	"sync"

	"lab1/internal/model"
)

// Generator: этап 1 (отдает посты в канал)
func Generate(posts []model.Post) <-chan model.Post {
	out := make(chan model.Post)
	go func() {
		for _, p := range posts {
			out <- p
		}
		close(out)
	}()
	return out
}

// Process: этап 2 (воркер, который обрабатывает данные)
func Process(in <-chan model.Post) <-chan model.Post {
	out := make(chan model.Post)
	go func() {
		for p := range in {
			// Имитация обработки: переводим заголовок в верхний регистр
			p.Title = strings.ToUpper(p.Title)
			out <- p
		}
		close(out)
	}()
	return out
}

// Merge: Fan-in (собирает данные из нескольких каналов-воркеров в один)
func Merge(cs ...<-chan model.Post) <-chan model.Post {
	var wg sync.WaitGroup
	out := make(chan model.Post)

	output := func(c <-chan model.Post) {
		for n := range c {
			out <- n
		}
		wg.Done()
	}

	wg.Add(len(cs))
	for _, c := range cs {
		go output(c)
	}

	// Ждем завершения всех горутин и закрываем итоговый канал
	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}
