TODO API

Установка и запуск
1. Клонируйте репозиторий
git clone https://github.com/theoreooo/todo
cd TODO

2. Запустите c помощью Docker

docker-compose up --build

API будет доступно по http://localhost:8082.

Остановка приложения:
docker-compose down -v

Запуск юнит тестов
Есть юнит тесты для хендлеров в internal/http-server/handlers. Они используют моки и не нуждаются в БД
1. Запуск тестов в Docker

docker-compose -f docker-compose.test.yml up --build test

Очистка:
docker-compose down -v


Документация API
Чтобы увидеть интерактивную документацию:

откройте в браузере http://localhost:8080.

Структура проекта
TODO/
├── cmd/
│   └── todo/
│       └── main.go         # Application entry point
├── config/
│   └── local.yaml         # Configuration file
├── internal/
│   ├── http-server/
│   │   └── handlers/      # HTTP handlers and unit tests
│   ├── models/            # Data models
│   └── storage/
│       └── postgres/      # PostgreSQL storage layer
├── docs/
│   └── openapi.yaml       # Swagger API documentation
├── docker-compose.yml     # Docker Compose configuration
├── Dockerfile             # Docker build instructions
└── go.mod                 # Go module dependencies

