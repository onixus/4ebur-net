# 🇷🇺 ALT Linux Support

4ebur-net теперь поддерживает сборку на базе **ALT Linux** — российской операционной системы с открытым исходным кодом.

## Быстрый старт

### ALT Sisyphus (Rolling Release)

```bash
# Сборка
docker build -f Dockerfile.alt -t 4ebur-net:alt-sisyphus .

# Запуск
docker run -d -p 1488:1488 4ebur-net:alt-sisyphus
```

### ALT P10 (Stable)

```bash
# Сборка
docker build -f Dockerfile.alt-p10 -t 4ebur-net:alt-p10 .

# Запуск
docker run -d -p 1488:1488 4ebur-net:alt-p10
```

### Docker Compose

```bash
# Запуск обеих версий
docker-compose -f docker-compose.alt.yml up -d
```

## Документация

Полная документация: [docs/ALT_LINUX.md](../docs/ALT_LINUX.md)

## Преимущества

✅ Российская разработка (импортозамещение)  
✅ Поддержка архитектур: x86_64, aarch64, Эльбрус, RISC-V  
✅ Открытый исходный код  
✅ Репозиторий Sisyphus (20K+ пакетов)  
✅ Production-ready (P10 stable branch)

## Сравнение

| Критерий | ALT Sisyphus | ALT P10 | Alpine | Scratch |
|----------|--------------|---------|--------|--------|
| **Размер** | ~180MB | ~200MB | ~25MB | ~15MB |
| **Go version** | 1.25.5 | 1.24.9 | 1.23 | N/A |
| **Российская ОС** | ✅ | ✅ | ❌ | ❌ |
| **Стабильность** | Rolling | ✅ Stable | ✅ | ✅ |
| **Use case** | Dev/Test | Production | Production | Micro |

## Ссылки

- [Официальный сайт ALT Linux](https://www.altlinux.org/)
- [Docker Hub - ALT образы](https://hub.docker.com/_/alt/)
- [Базальт СПО](https://www.basealt.ru/)
