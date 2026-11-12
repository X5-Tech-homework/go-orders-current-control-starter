# Сервис заказов

Простой REST API для создания заказов и изменения их статусов. Данные хранятся в памяти и исчезают после остановки приложения.

## Запуск

Нужен Go 1.22 или новее.

```bash
go mod download
go run ./cmd/api
```

По умолчанию сервер работает на порту 8080. Адрес можно изменить через переменную `HTTP_ADDR`, например `:8081`.

Swagger UI: http://localhost:8080/swagger/

OpenAPI JSON: http://localhost:8080/swagger/openapi.json

Swagger UI загружает JavaScript и CSS через CDN, поэтому для открытия интерфейса нужен интернет. Описание API находится в `docs/openapi.json` и встроено в приложение; оно написано вручную и не генерируется из аннотаций.

## API

| Метод | Путь | Что делает |
| --- | --- | --- |
| GET | `/health` | Проверяет доступность |
| POST | `/api/v1/orders` | Создает заказ |
| GET | `/api/v1/orders` | Возвращает список заказов |
| GET | `/api/v1/orders/{id}` | Возвращает заказ |
| PATCH | `/api/v1/orders/{id}/status` | Меняет статус |
| GET | `/api/v1/orders/{id}/history` | Возвращает историю статусов |

Пример создания заказа:

```json
{
  "items": [
    {"name": "Keyboard", "quantity": 1, "price": 500000}
  ]
}
```

Цена хранится в минимальных денежных единицах, например в копейках. Количество и цена должны быть больше нуля, название не должно быть пустым.

Пример изменения статуса:

```json
{"status": "processing"}
```

Новый заказ получает статус `new`. Разрешенные переходы: `new` в `processing` или `cancelled`, `processing` в `completed` или `cancelled`. Статусы `completed` и `cancelled` являются конечными.

Ошибки возвращаются в формате:

```json
{"error": "order not found", "code": "ORDER_NOT_FOUND"}
```

Для некорректных данных используется HTTP 400, для отсутствующего заказа 404, для недопустимого перехода 409, для внутренней ошибки 500.

## Проверка

```bash
go test ./...
go test ./internal/service/... -cover
go vet ./...
go build ./...
```

## Структура

- `cmd/api` - запуск HTTP-сервера.
- `internal/order` - модель, ошибки и интерфейс репозитория.
- `internal/repository/memory` - хранение заказов в памяти.
- `internal/service` - бизнес-логика.
- `internal/handler` - HTTP-обработчики и middleware.
- `docs` - OpenAPI и страница Swagger UI.
