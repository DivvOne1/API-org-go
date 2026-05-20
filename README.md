# API организационной структуры

Тестовое задание на Go: API для работы с подразделениями и сотрудниками.

Используемый стек:
- Go
- `net/http`
- GORM
- PostgreSQL
- goose
- Docker / Docker Compose

## Возможности

- создание подразделений;
- создание сотрудников в подразделении;
- получение подразделения вместе с деревом дочерних подразделений;
- перенос подразделения в другое подразделение;
- удаление подразделения в режимах `cascade` и `reassign`;
- валидация входных данных;
- защита от циклов при переносе подразделения;
- миграции через `goose`;
- запуск через `docker-compose`;
- базовые тесты.

## Структура проекта

```text
.
├── cmd/api/main.go
├── internal/app
├── internal/config
├── internal/db
├── internal/models
├── internal/repository
├── internal/service
├── internal/handler
├── internal/response
├── migrations
├── Dockerfile
├── docker-compose.yml
├── .env.example
└── README.md
```

## Запуск

```bash
docker-compose up --build
```

После запуска API будет доступно по адресу:

`http://localhost:8080`

Порт `5432` у PostgreSQL проброшен наружу только для удобства локальной разработки и ручной проверки базы.

## Эндпоинты

- `POST /departments`
- `POST /departments/{id}/employees`
- `GET /departments/{id}?depth=3&include_employees=true`
- `PATCH /departments/{id}`
- `DELETE /departments/{id}?mode=cascade`
- `DELETE /departments/{id}?mode=reassign&reassign_to_department_id=10`

Поддерживаются оба варианта URL: со слешем на конце и без него.

Параметр `mode` для `DELETE /departments/{id}` обязателен.

## Примеры запросов

Создать подразделение:

```bash
curl -X POST http://localhost:8080/departments \
  -H "Content-Type: application/json" \
  -d '{"name":"Backend"}'
```

Создать дочернее подразделение:

```bash
curl -X POST http://localhost:8080/departments \
  -H "Content-Type: application/json" \
  -d '{"name":"Platform","parent_id":1}'
```

Создать сотрудника:

```bash
curl -X POST http://localhost:8080/departments/1/employees \
  -H "Content-Type: application/json" \
  -d '{"full_name":"Ivan Petrov","position":"Go Developer","hired_at":"2025-11-01"}'
```

Получить дерево подразделений:

```bash
curl "http://localhost:8080/departments/1?depth=2&include_employees=true"
```

Переместить подразделение:

```bash
curl -X PATCH http://localhost:8080/departments/2 \
  -H "Content-Type: application/json" \
  -d '{"parent_id":1,"name":"Platform"}'
```

Удалить подразделение каскадно:

```bash
curl -X DELETE "http://localhost:8080/departments/2?mode=cascade"
```

Удалить подразделение с переводом сотрудников:

```bash
curl -X DELETE "http://localhost:8080/departments/2?mode=reassign&reassign_to_department_id=1"
```

## Основные бизнес-правила

- `name` у подразделения обязателен, обрезается по краям и ограничен 200 символами;
- `full_name` и `position` у сотрудника обязательны и ограничены 200 символами;
- в рамках одного `parent_id` названия подразделений уникальны;
- для корневых подразделений используется отдельный partial unique index, потому что в PostgreSQL `NULL != NULL`;
- нельзя сделать подразделение родителем самого себя;
- нельзя переместить подразделение внутрь собственного поддерева;
- `depth` по умолчанию равен `1`, допустимый диапазон: `0..5`;
- сотрудники в ответе сортируются по `full_name`.

## Поведение удаления

### `mode=cascade`

Удаляется всё поддерево подразделения вместе с сотрудниками.

Каскадное удаление обеспечивается внешними ключами с `ON DELETE CASCADE`.

### `mode=reassign`

Удаляется всё поддерево подразделения.

Перед удалением все сотрудники из всего поддерева переводятся в подразделение `reassign_to_department_id`.

Ограничения:

- `reassign_to_department_id` обязателен;
- нельзя переводить сотрудников в удаляемое подразделение;
- нельзя переводить сотрудников в подразделение из удаляемого поддерева.

## Миграции

Начальная миграция создаёт:

- `departments.parent_id` как FK на `departments.id`;
- partial unique indexes для корневых и вложенных подразделений;
- `employees.department_id` как FK на `departments.id`;
- каскадное удаление для сотрудников и дочерних подразделений.

## Тесты

Если Go установлен локально:

```bash
go test ./...
```

Проверка через Docker:

```bash
docker build -t org-structure-api-test .
```

В проекте есть тесты на:

- создание подразделения;
- защиту от цикла при `PATCH /departments/{id}`;
- возврат дерева подразделений;
- service-level проверку цикла при переносе подразделения.

## Примечания

- дерево подразделений собирается рекурсивно на уровне service;
- для production-решения это можно оптимизировать через recursive CTE;
- при гонке двух запросов на одинаковое имя подразделения конфликт по unique index также возвращается как `409 Conflict`.
