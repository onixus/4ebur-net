# Multi-stage build для минимального размера образа
# Stage 1: Сборка
FROM golang:1.21-alpine AS builder

# Устанавливаем необходимые инструменты
RUN apk add --no-cache git ca-certificates tzdata

# Устанавливаем рабочую директорию
WORKDIR /build

# Копируем go.mod и go.sum для кэширования зависимостей
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем приложение с оптимизациями для production
# -ldflags="-s -w" удаляет отладочную информацию и таблицу символов
# CGO_ENABLED=0 создает статический бинарник
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w -X main.Version=$(git describe --tags --always --dirty)" \
    -trimpath \
    -o 4ebur-net \
    cmd/proxy/main.go

# Stage 2: Минимальный runtime образ
FROM scratch

# Копируем CA сертификаты для работы с HTTPS
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Копируем информацию о временных зонах
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Копируем скомпилированный бинарник
COPY --from=builder /build/4ebur-net /4ebur-net

# Создаем непривилегированного пользователя
USER 65534:65534

# Открываем порт прокси
EXPOSE 8080

# Устанавливаем переменные окружения по умолчанию
ENV PROXY_PORT=8080 \
    MAX_IDLE_CONNS=1000 \
    MAX_IDLE_CONNS_PER_HOST=100 \
    MAX_CONNS_PER_HOST=100 \
    TZ=UTC

# Метаданные образа
LABEL maintainer="onixus <onixus@live.ru>" \
      description="High-performance MITM forward proxy" \
      version="1.0.0"

# Healthcheck для мониторинга
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/4ebur-net", "--health"]

# Запускаем прокси
ENTRYPOINT ["/4ebur-net"]
