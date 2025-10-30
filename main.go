package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Структура для конфигурации
type Config struct {
	ServerPort      string
	LauncherClient  string
	GameClient      string
	LauncherVersion string
	GameVersion     string
	ClientsDir      string
}

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

type VersionResponse struct {
	LauncherVersion string `json:"launcher_version"`
	GameVersion     string `json:"game_version"`
}

type FileInfoResponse struct {
	Filename string `json:"filename"`
	Size     int64  `json:"size"`
	Hash     string `json:"hash"`
}

// Структура для логгера с дополнительными полями
type Logger struct {
	*log.Logger
}

var config Config

func main() {
	// Загружаем конфигурацию
	if err := loadConfig(); err != nil {
		log.Fatalf("❌ Ошибка загрузки конфигурации: %v", err)
	}

	// Создаем логгер с префиксом и датой
	logger := &Logger{
		Logger: log.New(os.Stdout, "[LAUNCHER] ", log.Ldate|log.Ltime),
	}

	// Статика для изображений
	http.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir("./images"))))

	// API эндпоинты с логированием
	http.HandleFunc("/api/news", logger.newsHandler)
	http.HandleFunc("/api/version", logger.versionHandler)
	http.HandleFunc("/api/download/launcher", logger.downloadLauncherHandler)
	http.HandleFunc("/api/download/game", logger.downloadGameHandler)

	// Запуск сервера
	port := ":" + config.ServerPort
	logger.Printf("Сервер лаунчера запущен на http://localhost%s", port)
	logger.Println("Готов к приему запросов...")
	log.Fatal(http.ListenAndServe(port, nil))
}

// Загрузка конфигурации из .env файла
func loadConfig() error {
	if err := godotenv.Load(); err != nil {
		return fmt.Errorf("ошибка загрузки .env файла: %v", err)
	}

	config = Config{
		ServerPort:      getEnv("SERVER_PORT", "8080"),
		LauncherClient:  getEnv("LAUNCHER_CLIENT_FILE", "launcher.exe"),
		GameClient:      getEnv("GAME_CLIENT_FILE", "Loil.exe"),
		LauncherVersion: getEnv("LAUNCHER_VERSION", "0.0.0"),
		GameVersion:     getEnv("GAME_VERSION", "0.0.0"),
		ClientsDir:      getEnv("CLIENTS_DIR", "clients"),
	}

	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Обработчик новостей с логированием
func (l *Logger) newsHandler(w http.ResponseWriter, r *http.Request) {
	l.handleWithCORS(w, r, "📰", "/api/news", func() {
		// Загружаем новости
		news, err := loadNews()
		if err != nil {
			l.logError("Ошибка загрузки новостей: %v", err)
			http.Error(w, fmt.Sprintf("Ошибка загрузки новостей: %v", err), http.StatusInternalServerError)
			return
		}

		// Отправляем ответ
		response := NewsResponse{News: news}
		json.NewEncoder(w).Encode(response)

		l.logSuccess("Отправлено новостей: %d", len(news))
	})
}

// Обработчик версий
func (l *Logger) versionHandler(w http.ResponseWriter, r *http.Request) {
	l.handleWithCORS(w, r, "🔖", "/api/version", func() {
		response := VersionResponse{
			LauncherVersion: config.LauncherVersion,
			GameVersion:     config.GameVersion,
		}

		json.NewEncoder(w).Encode(response)
		l.logSuccess("Отправлены версии: лаунчер=%s, игра=%s",
			config.LauncherVersion, config.GameVersion)
	})
}

// Обработчик скачивания лаунчера
func (l *Logger) downloadLauncherHandler(w http.ResponseWriter, r *http.Request) {
	l.handleWithCORS(w, r, "⬇️", "/api/download/launcher", func() {
		filePath := filepath.Join(config.ClientsDir, config.LauncherClient)
		l.serveFileDownload(w, r, filePath, "launcher")
	})
}

// Обработчик скачивания игры
func (l *Logger) downloadGameHandler(w http.ResponseWriter, r *http.Request) {
	l.handleWithCORS(w, r, "⬇️", "/api/download/game", func() {
		filePath := filepath.Join(config.ClientsDir, config.GameClient)
		l.serveFileDownload(w, r, filePath, "game")
	})
}

// Общая логика для скачивания файлов
func (l *Logger) serveFileDownload(w http.ResponseWriter, r *http.Request, filePath, fileType string) {
	// Проверяем существование файла
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		l.logError("Файл не найден: %s", filePath)
		http.Error(w, "Файл не найден", http.StatusNotFound)
		return
	}

	// Открываем файл
	file, err := os.Open(filePath)
	if err != nil {
		l.logError("Ошибка открытия файла %s: %v", filePath, err)
		http.Error(w, "Ошибка открытия файла", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// Получаем информацию о файле
	fileInfo, err := file.Stat()
	if err != nil {
		l.logError("Ошибка получения информации о файле %s: %v", filePath, err)
		http.Error(w, "Ошибка получения информации о файле", http.StatusInternalServerError)
		return
	}

	// Вычисляем хэш файла
	hash, err := calculateFileHash(filePath)
	if err != nil {
		l.logError("Ошибка вычисления хэша файла %s: %v", filePath, err)
		// Не прерываем выполнение, хэш не обязателен для скачивания
	}

	// Получаем только имя файла для заголовка
	filename := filepath.Base(filePath)

	// Устанавливаем заголовки
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))

	// Добавляем информацию о хэше в заголовок, если удалось вычислить
	if hash != "" {
		w.Header().Set("X-File-Hash", hash)
	}

	// Копируем файл в ответ
	_, err = io.Copy(w, file)
	if err != nil {
		l.logError("Ошибка отправки файла %s: %v", filePath, err)
		return
	}

	l.logSuccess("Отправлен файл %s (размер: %d bytes, хэш: %s)",
		filename, fileInfo.Size(), hash)
}

// Общая обработка CORS и логирования
func (l *Logger) handleWithCORS(w http.ResponseWriter, r *http.Request, emoji, endpoint string, handler func()) {
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
	l.Printf("%s Запрос %s от %s", emoji, endpoint, clientIP)

	// Выполняем основной обработчик
	handler()

	// Логируем в файл
	l.logToFile(clientIP, endpoint, emoji)
}

// Логирование ошибки
func (l *Logger) logError(format string, v ...interface{}) {
	message := fmt.Sprintf(format, v...)
	l.Printf("❌ %s", message)
}

// Логирование успеха
func (l *Logger) logSuccess(format string, v ...interface{}) {
	message := fmt.Sprintf(format, v...)
	l.Printf("✅ %s", message)
}

// Логирование в файл с датой
func (l *Logger) logToFile(clientIP, endpoint, emoji string) {
	date := time.Now().Format("2006-01-02")
	logDir := "logs"
	logFile := filepath.Join(logDir, fmt.Sprintf("access_%s.log", date))

	// Создаем директорию если не существует
	if err := os.MkdirAll(logDir, 0755); err != nil {
		l.Printf("❌ Ошибка создания директории логов: %v", err)
		return
	}

	// Открываем файл для добавления логов
	file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		l.Printf("❌ Ошибка открытия файла логов: %v", err)
		return
	}
	defer file.Close()

	logEntry := fmt.Sprintf("[%s] %s %s - %s\n",
		time.Now().Format("2006-01-02 15:04:05"),
		clientIP,
		endpoint,
		emoji)

	if _, err := file.WriteString(logEntry); err != nil {
		l.Printf("❌ Ошибка записи в файл логов: %v", err)
	}
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

// Вычисление хэша файла
func calculateFileHash(filename string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
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
