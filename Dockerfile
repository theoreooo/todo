FROM golang:1.23 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /todo-app ./cmd/todo/main.go

FROM golang:1.23 AS tester

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

CMD ["go", "test", "-v", "./internal/http-server/handlers"]

FROM alpine:latest

RUN apk --no-cache add ca-certificates

COPY --from=builder /todo-app /todo-app

COPY config/local.yaml /config.yaml

ENV CONFIG_PATH=/config.yaml

EXPOSE 8082

CMD ["/todo-app"]