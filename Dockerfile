FROM golang:1.21-alpine

WORKDIR /app

# Установка необходимых пакетов
RUN apk add --no-cache git

# Копирование go.mod и go.sum
COPY go.mod go.sum ./

# Загрузка зависимостей
RUN go mod download

# Команда по умолчанию
CMD ["sh"]