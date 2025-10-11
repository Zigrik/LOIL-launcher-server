package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

// Структура для новостей
type NewsItem struct {
	ID      int    `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
	Image   string `json:"image"` // имя JPG файла
	Date    string `json:"date"`
}

type NewsResponse struct {
	News []NewsItem `json:"news"`
}

func main() {
	// Статика для изображений
	http.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir("./images"))))

	// API для новостей
	http.HandleFunc("/api/news", newsHandler)

	// Запуск сервера
	port := ":8080"
	log.Printf("Сервер лаунчера запущен на http://localhost%s", port)
	log.Fatal(http.ListenAndServe(port, nil))
}

func newsHandler(w http.ResponseWriter, r *http.Request) {
	// Разрешаем CORS для Godot клиента
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	// Загружаем новости из JSON файла
	news, err := loadNews()
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка загрузки новостей: %v", err), http.StatusInternalServerError)
		return
	}

	// Отправляем клиенту
	response := NewsResponse{News: news}
	json.NewEncoder(w).Encode(response)
}

func loadNews() ([]NewsItem, error) {
	// Читаем JSON файл
	data, err := os.ReadFile("news/news.json")
	if err != nil {
		return nil, err
	}

	var news []NewsItem
	err = json.Unmarshal(data, &news)
	return news, err
}
