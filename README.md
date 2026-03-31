# GoCrawl

[![Tests](https://github.com/suprt/gocrawl/actions/workflows/test.yml/badge.svg)](https://github.com/suprt/gocrawl/actions/workflows/test.yml)

CLI-утилита для **параллельного скачивания веб-страниц** по списку URL с настраиваемым уровнем параллелизма, retry-логикой и прогресс-баром.

## 🚀 Возможности

- ⚡ **Параллельная загрузка** — настраиваемое количество воркеров (по умолчанию 10)
- 🔄 **Retry-логика** — экспоненциальная задержка при ошибках (до 5 попыток)
- 📊 **Progress bar** — визуальное отображение прогресса загрузки
- 🎛️ **Rate limiting** — контроль частоты запросов (мс между запросами)
- 🏷️ **Auto-detection** — автоматическое определение расширения по Content-Type
- 🔁 **Redirect limit** — ограничение на редиректы (до 10)
- 📁 **Пакетная загрузка** — загрузка из файла со списком URL
- 🛑 **Graceful shutdown** — корректная остановка по Ctrl+C (двойной Ctrl+C для немедленного выхода)
- 📝 **Структурированное логирование** — slog с уровнями DEBUG/INFO/WARN/ERROR
- 🐳 **Docker** — готовый multi-stage образ
- 🧪 **Тесты + бенчмарки** — покрытие тестами, benchmark для производительности

## 📦 Установка

### Из исходного кода

```bash
git clone https://github.com/suprt/gocrawl.git
cd gocrawl
go build -o gocrawl ./cmd/gocrawl
```

### Через Docker

```bash
docker build -t gocrawl .
```

## 💡 Использование

### Базовые примеры

```bash
# Скачать одну страницу
./gocrawl https://example.com

# Скачать несколько страниц
./gocrawl https://example.com https://golang.org https://google.com

# Скачать из файла со списком URL
./gocrawl -file urls.txt

# Скачать с кастомными настройками
./gocrawl -file urls.txt -workers 20 -timeout 60s -output ./downloads
```

### Все опции командной строки

| Опция | Описание | По умолчанию |
|-------|----------|--------------|
| `-file` | Файл со списком URL (один на строку) | — |
| `-output` | Директория для сохранения файлов | `./downloads` |
| `-workers` | Количество параллельных воркеров | `10` |
| `-timeout` | Таймаут для каждого запроса | `30s` |
| `-retries` | Максимальное количество попыток | `5` |
| `-progress` | Показывать progress bar | `true` |
| `-rate-limit` | Задержка между запросами (мс) | `0` |
| `-verbose` | Включить DEBUG-логирование | `false` |

### Переменные окружения

| Переменная | Описание | Приоритет |
|------------|----------|-----------|
| `GOCRAWL_WORKERS` | Количество воркеров | Выше флага |
| `GOCRAWL_TIMEOUT` | Таймаут (duration формат) | Выше флага |
| `GOCRAWL_RETRIES` | Количество попыток | Выше флага |
| `GOCRAWL_RATELIMIT` | Задержка в мс | Выше флага |
| `GOCRAWL_MAX_DURATION` | Макс. время работы (например, `10m`, `1h`) | Выше флага |

> **Примечание**: Переменные окружения имеют приоритет над флагами командной строки.

### Примеры использования

```bash
# Быстрая загрузка с 20 воркерами
./gocrawl -workers 20 -file urls.txt

# Осторожная загрузка с rate limiting (100ms между запросами)
./gocrawl -rate-limit 100 -file urls.txt

# Загрузка с таймаутом 2 минуты на запрос
./gocrawl -timeout 2m -file urls.txt

# Тихий режим (без progress bar, только ошибки)
./gocrawl -progress=false -file urls.txt

# Подробное логирование (для отладки)
./gocrawl -verbose -file urls.txt

# Ограничить время работы 30 минутами
GOCRAWL_MAX_DURATION=30m ./gocrawl -file urls.txt
```

### Формат файла со списком URL

```
https://example.com
https://golang.org
# Это комментарий - будет проигнорирован
https://google.com

# Пустые строки тоже игнорируются
```

## 🐳 Docker

### Сборка образа

```bash
docker build -t gocrawl .
```

### Запуск контейнера

```bash
# Linux/Mac
docker run --rm -v ${PWD}:/data gocrawl https://example.com -output /data/downloads

# Windows (PowerShell)
docker run --rm -v "${PWD}:/data" gocrawl https://example.com -output /data/downloads

# Скачать из файла
docker run --rm -v "${PWD}:/data" gocrawl -file /data/urls.txt -output /data/downloads
```

> **Важно**: Для доступа к файлам на хосте используйте `-v ${PWD}:/data` для монтирования тома.

## 🏗️ Архитектура

```
gocrawl/
├── cmd/gocrawl/           # Точка входа
│   └── main.go
├── internal/
│   ├── bootstrap/         # Инициализация приложения, DI
│   ├── config/            # Конфигурация (CLI флаги + env)
│   ├── crawler/           # Оркестрация воркеров, retry-логика
│   ├── downloader/        # HTTP-клиент, декодирование charset
│   ├── logger/            # Обёртка над log/slog
│   ├── naming/            # Генерация имён файлов
│   ├── parser/            # Парсинг и нормализация URL
│   ├── progress/          # Progress bar
│   └── storage/           # Сохранение файлов на диск
├── scripts/
│   └── build.ps1          # PowerShell скрипт для Windows
├── .github/workflows/
│   └── test.yml           # GitHub Actions (CI/CD)
├── Dockerfile             # Multi-stage сборка
├── Makefile               # Команды для разработки (Linux/Mac)
└── go.mod
```

## 🔧 Конкурентность

- **Воркеры**: настраиваемое количество горутин (по умолчанию 10)
- **Каналы**: `jobs`, `results`, `errors`, `retryJobs` для коммуникации
- **Graceful shutdown**: Отмена контекста по Ctrl+C, двойной Ctrl+C для немедленного выхода
- **Retry**: Экспоненциальная задержка (1s, 2s, 4s, 8s, ...)

## 📊 Расширения файлов

Утилита автоматически определяет расширение по Content-Type:

| Content-Type | Расширение |
|--------------|------------|
| `text/html` | `.html` |
| `text/css` | `.css` |
| `application/javascript` | `.js` |
| `application/json` | `.json` |
| `image/png` | `.png` |
| `image/jpeg` | `.jpg` |
| `application/pdf` | `.pdf` |
| `audio/*` | `.mp3` |
| `video/*` | `.mp4` |
| `text/xml`, `application/xml` | `.xml` |
| другие | `.bin` |

## 📝 Именование файлов

Файлы сохраняются в формате:
```
{readable-part}_{hash}.{ext}
```

| Компонент | Описание |
|-----------|----------|
| `readable-part` | Последний сегмент пути или хост |
| `hash` | Первые 8 символов MD5 от URL |
| `ext` | Расширение по Content-Type |

**Примеры**:
- `https://example.com` → `example_a1b2c3d4.html`
- `https://golang.org/doc/install` → `install_e5f6g7h8.html`

## 🧪 Тестирование

```bash
# Все тесты
go test ./...

# Тесты с выводом
go test -v ./...

# Тесты с покрытием
go test -cover ./...

# Тесты с race detector
go test -race ./...

# Бенчмарки
go test -bench=. -benchmem ./internal/crawler
go test -bench=. -benchmem ./internal/downloader
```

### Покрытие тестами

| Пакет | Тесты | Бенчмарки |
|-------|-------|-----------|
| `crawler` | ✅ 7 тестов | ✅ 10 бенчмарков |
| `downloader` | ✅ 6 тестов | ✅ 11 бенчмарков |
| `naming` | ✅ 5 тестов | — |
| `parser` | ✅ 4 теста | — |
| `storage` | ✅ 8 тестов | — |

## 🛠️ Разработка

### Требования

- Go 1.25.4+
- Docker (опционально)
- Make (опционально)

### Makefile (Linux/Mac)

```bash
make build          # Сборка бинарника
make test           # Запуск тестов
make test-race      # Тесты с race detector
make lint           # Запуск линтера
make run ARGS='...' # Запуск приложения
make docker-build   # Сборка Docker образа
make docker-run ARGS='...' # Запуск Docker контейнера
make clean          # Очистка
```

### PowerShell (Windows)

```powershell
.\scripts\build.ps1 -Command build       # Сборка
.\scripts\build.ps1 -Command test        # Тесты
.\scripts\build.ps1 -Command test-race   # Тесты с race detector
.\scripts\build.ps1 -Command lint        # Линтер
.\scripts\build.ps1 -Command run -Args '-file urls.txt'  # Запуск
.\scripts\build.ps1 -Command docker-build  # Docker образ
.\scripts\build.ps1 -Command clean       # Очистка
```

### CI/CD

Проект использует **GitHub Actions** для автоматического тестирования:

- ✅ Запуск тестов при каждом push/PR
- ✅ Race detector
- ✅ go vet
- ✅ golangci-lint

Workflow: `.github/workflows/test.yml`

## 📝 Логирование

### Уровни логирования

| Уровень | Когда |
|---------|-------|
| `DEBUG` | С флагом `-verbose` |
| `INFO` | Успешные операции |
| `WARN` | Предупреждения |
| `ERROR` | Ошибки |

### Примеры вывода

**Обычный режим**:
```
time=2026-03-30T09:00:00Z level=INFO msg="starting crawler" workers=10 retries=5 timeout=30s
time=2026-03-30T09:00:01Z level=INFO msg="Crawler finished" success=3 failed=1
time=2026-03-30T09:00:01Z level=ERROR msg="Job failed" error="job 2 failed: timeout" count=1
```

**Подробный режим** (`-verbose`):
```
time=2026-03-30T09:00:00Z level=DEBUG msg="Parsed URLs" count=4
time=2026-03-30T09:00:00Z level=DEBUG msg="URL" index=0 url="https://example.com"
time=2026-03-30T09:00:01Z level=DEBUG msg="Downloaded" url="https://example.com" path="downloads/example.html" status=200
```
```
