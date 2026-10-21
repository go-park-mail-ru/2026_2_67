# --- ЭТАП 1: Сборка приложения ---
FROM golang:1.27-alpine AS builder

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем файлы зависимостей
COPY go.mod go.sum ./
RUN go mod download

# Копируем ВСЕ файлы проекта в контейнер
COPY . .

# Собираем бинарник НАПРЯМУЮ через go build, без Makefile.
# Мы указываем путь ./backend/cmd/server, так как папка backend лежит в корне.
RUN go build -o ./backend/bin/server ./backend/cmd/server

# --- ЭТАП 2: Финальный образ ---
FROM alpine:latest

WORKDIR /root/

# Копируем бинарник из папки, куда мы его только что собрали
COPY --from=builder /app/backend/bin/server ./server
COPY .env .env  

EXPOSE 8080

CMD ["./server"]
