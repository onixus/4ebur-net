# 4ebur-net на ALT Linux 🇷🇺

🚀 **Полная поддержка российской операционной системы ALT Linux**

## 🎯 Поддерживаемые версии

| Версия ALT Linux | Docker образ | Go версия | Статус |
|------------------|--------------|-----------|--------|
| **Sisyphus** (rolling) | `4ebur-net:alt-sisyphus` | 1.25+ | ✅ Полная поддержка |
| **P10** (stable) | `4ebur-net:alt-p10` | 1.24+ | ✅ Полная поддержка |
| **P9** (legacy) | N/A | 1.16 | ⚠️ Устарело |

## 📦 Установка

### Вариант 1: Docker (рекомендуется)

#### Сборка образа

```bash
# Клонировать репозиторий
git clone https://github.com/onixus/4ebur-net.git
cd 4ebur-net

# Собрать Docker образ на базе ALT Sisyphus
docker build -f Dockerfile.alt -t 4ebur-net:alt .

# Или использовать docker-compose
docker-compose -f docker-compose.alt.yml build
```

#### Запуск контейнера

```bash
# Простой запуск
docker run -d \
  --name 4ebur-net-alt \
  -p 1488:1488 \
  4ebur-net:alt

# С настройками производительности
docker run -d \
  --name 4ebur-net-alt \
  -p 1488:1488 \
  -e CACHE_SIZE_MB=500 \
  -e CACHE_MAX_AGE=15m \
  -e MAX_IDLE_CONNS=2000 \
  --restart unless-stopped \
  4ebur-net:alt

# Используя docker-compose
docker-compose -f docker-compose.alt.yml up -d
```

#### Проверка работы

```bash
# Проверить статус контейнера
docker ps | grep 4ebur-net

# Посмотреть логи
docker logs -f 4ebur-net-alt

# Проверить веб-интерфейс
curl http://localhost:1488/health
```

### Вариант 2: Нативная сборка на ALT Linux

#### Установка зависимостей

**ALT Sisyphus:**

```bash
# Обновить систему
sudo apt-get update
sudo apt-get dist-upgrade

# Установить Go и git
sudo apt-get install -y golang git make

# Проверить версию Go
go version  # Должно быть >= 1.21
```

**ALT P10:**

```bash
# Обновить систему
sudo apt-get update
sudo apt-get upgrade

# Установить Go
sudo apt-get install -y golang git make

# Если версия Go старая, установить из бинарников
wget https://go.dev/dl/go1.24.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.24.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
```

#### Сборка из исходников

```bash
# Клонировать репозиторий
git clone https://github.com/onixus/4ebur-net.git
cd 4ebur-net

# Загрузить зависимости
go mod download

# Собрать бинарник
make build

# Или вручную
go build -o 4ebur-net ./cmd/proxy

# Проверить сборку
./4ebur-net --help
```

## 🔒 Установка CA сертификата на ALT Linux

### Скачивание сертификата

```bash
# Скачать CA сертификат из прокси
curl http://localhost:1488/ca.crt -o 4ebur-net-ca.crt

# Проверить содержимое
openssl x509 -in 4ebur-net-ca.crt -text -noout | grep -A 2 "Subject:"
```

### Установка в системное хранилище

**ALT Sisyphus / P10:**

```bash
# Скопировать в системную директорию
sudo cp 4ebur-net-ca.crt /usr/share/ca-certificates/

# Обновить хранилище сертификатов
sudo update-ca-certificates

# Или альтернативный способ
sudo cp 4ebur-net-ca.crt /etc/pki/ca-trust/source/anchors/
sudo update-ca-trust

# Проверить установку
trust list | grep -i "4ebur"
```

## ⚙️ Настройка прокси в ALT Linux

### Системные настройки (GNOME)

```bash
# Установить прокси через gsettings
gsettings set org.gnome.system.proxy mode 'manual'
gsettings set org.gnome.system.proxy.http host 'localhost'
gsettings set org.gnome.system.proxy.http port 1488
gsettings set org.gnome.system.proxy.https host 'localhost'
gsettings set org.gnome.system.proxy.https port 1488

# Отключить прокси
gsettings set org.gnome.system.proxy mode 'none'
```

### Переменные окружения

```bash
# Добавить в ~/.bashrc или ~/.zshrc
echo 'export HTTP_PROXY="http://localhost:1488"' >> ~/.bashrc
echo 'export HTTPS_PROXY="http://localhost:1488"' >> ~/.bashrc
echo 'export NO_PROXY="localhost,127.0.0.1,.local"' >> ~/.bashrc

# Применить изменения
source ~/.bashrc

# Проверить
env | grep -i proxy
```

## 📊 Мониторинг и отладка

### Просмотр статистики

```bash
# Статистика кеша
curl -s http://localhost:1488/stats | jq .

# Health check
curl -s http://localhost:1488/health | jq .

# В реальном времени
watch -n 1 'curl -s http://localhost:1488/stats | jq .'
```

### Логи Docker

```bash
# Все логи
docker logs 4ebur-net-alt

# Последние 50 строк
docker logs --tail 50 4ebur-net-alt

# В реальном времени
docker logs -f 4ebur-net-alt

# Только ошибки
docker logs 4ebur-net-alt 2>&1 | grep "✗"

# Только Cache операции
docker logs 4ebur-net-alt 2>&1 | grep -E "Cache|HIT|MISS"
```

## 🔧 Решение проблем

### Прокси не запускается

```bash
# Проверить порт занят ли
sudo netstat -tulpn | grep 1488
sudo ss -tulpn | grep 1488

# Убить процесс на порту
sudo fuser -k 1488/tcp

# Проверить логи
docker logs 4ebur-net-alt
```

### Ошибки сертификатов

```bash
# Переустановить CA сертификат
curl http://localhost:1488/ca.crt -o 4ebur-net-ca.crt
sudo cp 4ebur-net-ca.crt /etc/pki/ca-trust/source/anchors/
sudo update-ca-trust

# Проверить установлен ли
trust list | grep -i 4ebur

# Проверить валидность сертификата
openssl verify 4ebur-net-ca.crt
```

## 📝 Примеры использования

### Использование с curl

```bash
# Простой запрос
curl -x http://localhost:1488 https://www.ya.ru

# С выводом заголовков
curl -v -x http://localhost:1488 https://www.ya.ru

# API запрос
curl -x http://localhost:1488 https://api.github.com/users/octocat | jq .

# Скачивание файла
curl -x http://localhost:1488 -O https://example.com/file.tar.gz
```

### Использование с git

```bash
# Настроить прокси для git
git config --global http.proxy http://localhost:1488
git config --global https.proxy http://localhost:1488

# Клонировать репозиторий
git clone https://github.com/onixus/4ebur-net.git

# Отключить прокси
git config --global --unset http.proxy
git config --global --unset https.proxy
```

## 🎉 Заключение

4ebur-net полностью совместим с ALT Linux и обеспечивает:

✅ **Нативная поддержка** - Сборка на базе официальных образов ALT Linux  
✅ **Высокая производительность** - 40-150x ускорение с кешированием  
✅ **Простота развёртывания** - Docker или systemd на выбор  
✅ **Безопасность** - Полная интеграция с системными сертификатами  
✅ **Мониторинг** - Встроенные метрики и health checks  

---

**💬 Вопросы? [Открыть issue](https://github.com/onixus/4ebur-net/issues)**

**⭐ Проект полезен? [Поставьте звезду!](https://github.com/onixus/4ebur-net)**
