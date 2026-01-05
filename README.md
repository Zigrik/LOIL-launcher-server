🚀 Лаунчер LOIL v 0.2.1
📌 Базовый URL
text
http://localhost:8080
(порт настраивается в .env файле через SERVER_PORT)

🔧 API ЭНДПОИНТЫ
📰 1. ПОЛУЧИТЬ НОВОСТИ
Запрос:

text
GET /api/news
Ответ:

json
{
  "news": [
    {
      "id": 1,
      "title": "Обновление лаунчера",
      "content": "Исправлены ошибки, улучшена производительность",
      "image": "update_v1.2.jpg",
      "date": "2024-01-15"
    }
  ]
}
Описание:

Возвращает массив новостей из файла news/news.json

Изображения доступны по пути /images/{имя_файла_из_image}

Формат даты: YYYY-MM-DD

🔖 2. ПОЛУЧИТЬ ВЕРСИИ
Запрос:

text
GET /api/version
Ответ:

json
{
  "launcher_version": "1.0.0",
  "game_version": "2.5.1"
}
Описание:

Версии берутся из .env файла

LAUNCHER_VERSION - версия лаунчера

GAME_VERSION - версия игры

⬇️ 3. СКАЧАТЬ ЛАУНЧЕР
Запрос:

text
GET /api/download/launcher
Заголовки ответа:

text
Content-Disposition: attachment; filename=launcher.exe
Content-Type: application/octet-stream
Content-Length: 5242880
X-File-Hash: d41d8cd98f00b204e9800998ecf8427e
Описание:

Файл берется из папки clients/

Имя файла настраивается в .env как LAUNCHER_CLIENT_FILE

В заголовке X-File-Hash передается MD5 хеш для проверки целостности

⬇️ 4. СКАЧАТЬ ИГРУ
Запрос:

text
GET /api/download/game
Заголовки ответа:

text
Content-Disposition: attachment; filename=Loil.exe
Content-Type: application/octet-stream
Content-Length: 104857600
X-File-Hash: 5d41402abc4b2a76b9719d911017c592
Описание:

Файл берется из папки clients/

Имя файла настраивается в .env как GAME_CLIENT_FILE

В заголовке X-File-Hash передается MD5 хеш

⚙️ КОНФИГУРАЦИЯ (.env файл)
env
SERVER_PORT=8080                    # Порт сервера
LAUNCHER_CLIENT_FILE=launcher.exe   # Имя файла лаунчера
GAME_CLIENT_FILE=Loil.exe          # Имя файла игры
LAUNCHER_VERSION=1.0.0              # Версия лаунчера
GAME_VERSION=2.5.1                  # Версия игры
CLIENTS_DIR=clients                 # Папка с файлами клиентов
📁 СТРУКТУРА ФАЙЛОВ СЕРВЕРА
text
├── main.go                          # Основной файл сервера
├── .env                            # Конфигурация
├── news/
│   └── news.json                   # JSON с новостями
├── images/
│   ├── update_v1.2.jpg             # Изображения для новостей
│   └── ...
├── clients/
│   ├── launcher.exe                # Файл лаунчера
│   └── Loil.exe                   # Файл игры
└── logs/
    └── access_2024-01-15.log       # Логи запросов
📋 ФОРМАТ ФАЙЛА НОВОСТЕЙ (news/news.json)
json
[
  {
    "id": 1,
    "title": "Новый патч",
    "content": "Добавлены новые карты",
    "image": "patch_2.5.jpg",      # Файл должен быть в папке /images/
    "date": "2024-01-15"
  },
  {
    "id": 2,
    "title": "Обновление безопасности",
    "content": "Улучшена защита",
    "image": "security_update.jpg",
    "date": "2024-01-10"
  }
]
🛠️ КОМАНДЫ ДЛЯ ТЕСТИРОВАНИЯ
bash
# 1. Получить новости (форматированный вывод)
curl -s http://localhost:8080/api/news | python3 -m json.tool

# 2. Получить версии
curl http://localhost:8080/api/version

# 3. Скачать лаунчер (сохранить с оригинальным именем)
curl -OJ http://localhost:8080/api/download/launcher

# 4. Скачать игру
curl -OJ http://localhost:8080/api/download/game

# 5. Проверить хеш скачанного файла
md5sum launcher.exe

# 6. Получить только заголовки
curl -I http://localhost:8080/api/download/launcher

# 7. Скачать изображение новости
curl -O http://localhost:8080/images/update_v1.2.jpg
🚀 ЗАПУСК СЕРВЕРА
Установите зависимости:

bash
go mod init launcher-server
go mod tidy
Настройте .env файл: (пример выше)

Создайте необходимые папки:

bash
mkdir -p news images clients logs
Добавьте файлы:

Поместите launcher.exe и Loil.exe в папку clients/

Создайте news.json в папке news/

Добавьте изображения в папку images/

Запустите сервер:

bash
go run main.go
Сервер запустится с сообщением:

text
[LAUNCHER] 2024/01/15 14:30:00 Сервер лаунчера запущен на http://localhost:8080
[LAUNCHER] 2024/01/15 14:30:00 Готов к приему запросов...
📊 ЛОГИРОВАНИЕ
Сервер логирует все запросы:

В консоль: с эмодзи и деталями

В файлы: в папке logs/access_YYYY-MM-DD.log

Пример лога:

text
[2024-01-15 14:35:22] 192.168.1.100 /api/news - 📰
[2024-01-15 14:35:25] 192.168.1.100 /api/download/launcher - ⬇️
🔒 CORS ПОЛИТИКА
Сервер поддерживает CORS:

Разрешены запросы с любых доменов (Access-Control-Allow-Origin: *)

Разрешены методы: GET, OPTIONS

Поддерживаются preflight-запросы

⚠️ ВАЖНЫЕ ЗАМЕЧАНИЯ
Все API эндпоинты поддерживают метод GET

Для скачивания файлов используйте -OJ флаг в curl для сохранения оригинального имени

Хеш файла передается в заголовке X-File-Hash

Изображения новостей должны быть в формате JPG

Логи автоматически ротируются по датам

Сервер проверяет существование файлов перед отправкой