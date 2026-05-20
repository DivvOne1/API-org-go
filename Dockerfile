FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .

RUN go build -o /app/bin/server ./cmd/api

RUN go install github.com/pressly/goose/v3/cmd/goose@v3.26.0

FROM alpine:3.22

WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY --from=builder /go/bin/goose /usr/local/bin/goose
COPY --from=builder /app/bin/server /app/bin/server
COPY --from=builder /app/migrations /app/migrations

EXPOSE 8080

CMD ["sh", "-c", "goose -dir /app/migrations postgres \"$DATABASE_URL\" up && /app/bin/server"]
