# TODO API

## Установка и запуск
1. Клонируйте репозиторий
```bash
git clone https://github.com/theoreooo/todo
```

2. Запустите c помощью Docker
```bash
docker-compose up --build
```
API будет доступно по http://localhost:8082.

Остановка приложения:
```bash
docker-compose down -v
```

## Запуск юнит тестов
Есть юнит тесты для хендлеров в internal/http-server/handlers. Они используют моки и не нуждаются в БД
1. Запуск тестов в Docker
```bash
docker-compose -f docker-compose.test.yml up --build test
```
Очистка:
```bash
docker-compose down -v
```

Документация API
Чтобы увидеть интерактивную документацию:

откройте в браузере http://localhost:8080.
