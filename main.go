package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
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

// Структура для логгера с дополнительными полями
type Logger struct {
	*log.Logger
}

func main() {
	// Создаем логгер с префиксом и датой
	logger := &Logger{
		Logger: log.New(os.Stdout, "[LAUNCHER] ", log.Ldate|log.Ltime),
	}

	// Статика для изображений
	http.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir("./images"))))

	// API для новостей с логированием
	http.HandleFunc("/api/news", logger.newsHandler)

	// Запуск сервера
	port := ":8080"
	logger.Printf("Сервер лаунчера запущен на http://localhost%s", port)
	logger.Println("Готов к приему запросов...")
	log.Fatal(http.ListenAndServe(port, nil))
}

// Обработчик новостей с логированием
func (l *Logger) newsHandler(w http.ResponseWriter, r *http.Request) {
	// Явно разрешаем CORS
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Обрабатываем preflight OPTIONS запрос
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Логируем запрос
	clientIP := getClientIP(r)
	l.Printf("📰 Запрос /api/news от %s", clientIP)

	// Загружаем новости
	news, err := loadNews()
	if err != nil {
		l.Printf("❌ Ошибка загрузки новостей: %v", err)
		http.Error(w, fmt.Sprintf("Ошибка загрузки новостей: %v", err), http.StatusInternalServerError)
		return
	}

	// Отправляем ответ
	response := NewsResponse{News: news}
	json.NewEncoder(w).Encode(response)

	l.Printf("✅ Отправлено новостей: %d для %s", len(news), clientIP)
}

// Функция для получения реального IP клиента
func getClientIP(r *http.Request) string {
	// Пробуем получить IP из заголовков (если за прокси/балансировщиком)
	ip := r.Header.Get("X-Real-IP")
	if ip == "" {
		ip = r.Header.Get("X-Forwarded-For")
		if ip != "" {
			// Берем первый IP из списка
			ips := strings.Split(ip, ",")
			ip = strings.TrimSpace(ips[0])
		}
	}

	// Если в заголовках нет, берем RemoteAddr
	if ip == "" {
		ip, _, _ = net.SplitHostPort(r.RemoteAddr)
	}

	return ip
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
