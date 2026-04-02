package main

import (
	"flag"
	"fmt"
	"log"

	"lab1/internal/api"
	"lab1/internal/pipeline"
)

func main() {
	// Флаг для ограничения вывода (по умолчанию 5)
	limit := flag.Int("limit", 5, "Количество выводимых постов")
	flag.Parse()

	// 1. Получаем данные с API
	posts, err := api.FetchPosts()
	if err != nil {
		log.Fatalf("Ошибка получения данных: %v", err)
	}

	// 2. Создаем генератор данных (начало пайплайна)
	gen := pipeline.Generate(posts)

	// 3. Fan-out: распределяем задачи на 3 горутины-воркера
	worker1 := pipeline.Process(gen)
	worker2 := pipeline.Process(gen)
	worker3 := pipeline.Process(gen)

	// 4. Fan-in: собираем результаты от всех воркеров
	out := pipeline.Merge(worker1, worker2, worker3)

	// 5. Вывод результата в stdout с учетом лимита
	count := 0
	for p := range out {
		fmt.Printf("[%d] Title: %s\n", p.ID, p.Title)
		count++
		if count >= *limit {
			break
		}
	}
}
