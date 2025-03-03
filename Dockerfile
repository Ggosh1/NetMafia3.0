# Этап сборки: используем официальный образ Go для компиляции приложения
FROM golang:1.20-alpine AS builder

# Отключаем cgo на уровне окружения
ENV CGO_ENABLED=0

WORKDIR /app

# Если у вас есть файлы go.mod и go.sum, копируем их и загружаем зависимости
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код проекта
COPY . .

# Компилируем приложение (получим бинарный файл mafia_game)
RUN GOOS=linux go build -a -v -o mafia_game .

# Финальный образ: минимальный Alpine
FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Копируем скомпилированный бинарник из этапа сборки
COPY --from=builder /app/mafia_game .

# Подготавливаем директорию для фронтенда, так как сервер ожидает файлы в ./frontend
RUN mkdir -p frontend

COPY --from=builder /app/frontend ./frontend

# Открываем порт, на котором работает приложение
EXPOSE 8080

# Команда для запуска приложения
CMD ["./mafia_game"]
