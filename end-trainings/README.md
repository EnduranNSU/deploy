# Trainings Service

Микросервис для управления информацией о пользователях.

## Требования

- Go 1.24+ (при запуске без Docker)
- Docker и Docker Compose (при запуске через Docker)

## Конфигурация

Перед запуском создайте config/config.yaml по примеру из config/config-example.yaml, но только при запуске через docker делаем порт строго 8080

## Обычный запуск
```bash
make build && make run
```
## Запуск через Docker
```bash
docker compose up
```

## Документация API
При запуске сервиса по пути /swagger/index.html или в папке docs