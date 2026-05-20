# Organizational Structure API

Test assignment service for departments and employees.

Stack:
- Go
- `net/http`
- GORM
- PostgreSQL
- goose migrations
- Docker / Docker Compose

## Features

- CRUD-style endpoints for departments and employees
- Nested department tree with configurable `depth`
- Optional employee inclusion via `include_employees`
- Validation for required fields and max length
- Trimmed department and employee string fields
- Duplicate department name protection within the same parent
- Cycle protection when moving a department
- Cascade delete via PostgreSQL foreign keys
- `reassign` delete mode for the whole department subtree
- Basic tests

## Project structure

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

## Run

```bash
docker-compose up --build
```

API will be available at `http://localhost:8080`.

## Endpoints

- `POST /departments`
- `POST /departments/{id}/employees`
- `GET /departments/{id}?depth=3&include_employees=true`
- `PATCH /departments/{id}`
- `DELETE /departments/{id}?mode=cascade`
- `DELETE /departments/{id}?mode=reassign&reassign_to_department_id=10`

Both `/departments` and `/departments/` work.

## Request examples

Create department:

```bash
curl -X POST http://localhost:8080/departments \
  -H "Content-Type: application/json" \
  -d '{"name":"Backend"}'
```

Create child department:

```bash
curl -X POST http://localhost:8080/departments \
  -H "Content-Type: application/json" \
  -d '{"name":"Platform","parent_id":1}'
```

Create employee:

```bash
curl -X POST http://localhost:8080/departments/1/employees \
  -H "Content-Type: application/json" \
  -d '{"full_name":"Ivan Petrov","position":"Go Developer","hired_at":"2025-11-01"}'
```

Get tree:

```bash
curl "http://localhost:8080/departments/1?depth=2&include_employees=true"
```

Move department:

```bash
curl -X PATCH http://localhost:8080/departments/2 \
  -H "Content-Type: application/json" \
  -d '{"parent_id":1,"name":"Platform"}'
```

Delete with cascade:

```bash
curl -X DELETE "http://localhost:8080/departments/2?mode=cascade"
```

Delete with reassign:

```bash
curl -X DELETE "http://localhost:8080/departments/2?mode=reassign&reassign_to_department_id=1"
```

## Business rules

- Department `name`, employee `full_name`, and employee `position` are required and limited to 200 characters.
- Department names are unique within the same parent.
- Root departments use a separate partial unique index because PostgreSQL allows multiple `NULL` values in a composite unique index.
- A department cannot become its own parent.
- A department cannot be moved inside its own subtree.
- `depth` defaults to `1` and is limited to `0..5`.
- Employees are sorted by `full_name`.

### Reassign mode

`mode=reassign` deletes the entire subtree of the selected department.

Before deletion, all employees from the whole subtree are moved to `reassign_to_department_id`.

The target department cannot be the department being deleted or any department inside that subtree.

## Migrations

The initial migration creates:

- `departments.parent_id` foreign key to `departments.id`
- partial unique indexes for root and nested department names
- `employees.department_id` foreign key to `departments.id`
- `ON DELETE CASCADE` for child departments and employees

## Tests

Run locally if Go is installed:

```bash
go test ./...
```

Or run in Docker:

```bash
docker build -t org-structure-api-test .
```

The project includes tests for:

- creating a department
- cycle protection on `PATCH /departments/{id}`
- returning a department tree

## Notes

- The tree is built recursively in the service layer, which is acceptable for a test assignment.
- For a production version, tree loading could be optimized with a recursive CTE.
