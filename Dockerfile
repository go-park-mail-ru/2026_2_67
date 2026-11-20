FROM golang:1.27-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o ./backend/bin/server ./backend/cmd/server

FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/backend/bin/server ./server
COPY .env .env  

ARG APP_PORT=8080
EXPOSE ${APP_PORT}

CMD ["./server"]
