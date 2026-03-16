# GoCrawl

CLI утилита для параллельного скачивания веб-страниц по списку URL.

## Возможности

- ✅ Параллельное скачивание с настраиваемым количеством воркеров
- ✅ Автоматические повторные попытки при ошибках (retry)
- ✅ Progress bar для отслеживания прогресса
- ✅ Rate limiting для контроля частоты запросов
- ✅ Определение типа контента и сохранение с правильным расширением
- ✅ Лимит редиректов
- ✅ Поддержка файла со списком URL
- ✅ Обработка Ctrl+C для graceful shutdown
- ✅ Unit-тесты

## Установка

```bash
go build -o gocrawl ./cmd/gocrawl
```

## Использование

### Базовое использование

```bash
# Скачать один URL
./gocrawl https://example.com

# Скачать несколько URL
./gocrawl https://example.com https://golang.org https://google.com

# Скачать из файла со списком URL
./gocrawl -file urls.txt
```

### Опции командной строки

| Опция | Описание | По умолчанию |
|-------|----------|--------------|
| `-file` | Файл со списком URL (один на строку) | `` |
| `-output` | Директория для сохранения файлов | `./downloads` |
| `-workers` | Количество параллельных воркеров | `5` |
| `-timeout` | Таймаут для каждого запроса | `30s` |
| `-retries` | Максимальное количество повторных попыток | `5` |
| `-progress` | Показывать progress bar | `true` |
| `-rate-limit` | Задержка между запросами (мс) | `0` |
| `-verbose` | Включить подробное логирование | `false` |

### Примеры

```bash
# Скачать с 10 воркерами и таймаутом 60 секунд
./gocrawl -workers 10 -timeout 60s https://example.com

# Скачать в другую директорию
./gocrawl -output ./my_downloads https://example.com

# Скачать с rate limiting (100ms между запросами)
./gocrawl -rate-limit 100 -file urls.txt

# Отключить progress bar
./gocrawl -progress=false https://example.com

# Включить подробное логирование (debug/info/error)
./gocrawl -verbose -file urls.txt

# Логирование только ошибок (по умолчанию)
./gocrawl -progress=false -file urls.txt
```

### Переменные окружения

| Переменная | Описание |
|------------|----------|
| `GOCRAWL_WORKERS` | Количество воркеров (переопределяет флаг) |
| `GOCRAWL_TIMEOUT` | Таймаут в формате duration (переопределяет флаг) |
| `GOCRAWL_RETRIES` | Количество повторных попыток (переопределяет флаг) |

## Формат файла со списком URL

```
https://example.com
https://golang.org
# Это комментарий - будет проигнорирован
https://google.com

# Пустые строки тоже игнорируются
```

## Архитектура

```
cmd/gocrawl/
└── main.go              # Точка входа

internal/
├── bootstrap/           # Инициализация приложения
├── config/              # Загрузка конфигурации (CLI флаги + env)
├── crawler/             # Оркестрация воркеров, retry-логика
├── downloader/          # HTTP-клиент с декодированием кодировок
├── naming/              # Генерация имён файлов (SHA256 + readable part)
├── parser/              # Парсинг и нормализация URL
├── progress/            # Progress bar
└── storage/             # Сохранение файлов
```

## Расширения файлов

Утилита автоматически определяет расширение по Content-Type:

| Content-Type | Расширение |
|--------------|------------|
| text/html | .html |
| text/css | .css |
| application/javascript | .js |
| application/json | .json |
| image/png | .png |
| image/jpeg | .jpg |
| application/pdf | .pdf |
| audio/* | .mp3 |
| video/* | .mp4 |
| text/xml, application/xml | .xml |
| другие | .bin |

## Именование файлов

Файлы сохраняются с именами вида:
```
{readable-part}_{hash}.{ext}
```

Где:
- `readable-part` — последний сегмент пути или имя хоста
- `hash` — первые 8 символов MD5 хэша полного URL
- `ext` — расширение по Content-Type

Примеры:
- `https://example.com` → `example_a1b2c3d4.html`
- `https://golang.org/doc/install` → `install_e5f6g7h8.html`

## Тесты

```bash
# Запустить все тесты
go test ./...

# Запустить тесты с выводом
go test ./... -v

# Запустить тесты с покрытием
go test ./... -cover
```

## Логирование

Утилита поддерживает два режима логирования:

### Обычный режим (по умолчанию)
Выводятся только ошибки и итоговая статистика:
```
Statistics:
  Success: 3
  Failed:  1

Error summary:
  job 3 failed after 6 attempts: ...
```

### Подробный режим (-verbose)
Выводится информация о каждом запросе, попытках retry и результатах:
```
Starting crawler with 5 workers, 5 retries, timeout=30s
Parsed 4 URL(s)
[DEBUG] Downloading https://example.com (attempt 1)
[DEBUG] Retry attempt 2/6 for https://invalid.com (waiting 1s)
[INFO] Downloaded https://example.com → downloads/example.html (200)
[ERROR] All retries exhausted for https://invalid.com after 6 attempts
```

Уровни логирования:
- `[DEBUG]` — отладочная информация (только с `-verbose`)
- `[INFO]` — успешные операции (всегда)
- `[ERROR]` — ошибки (всегда)

## Сборка и запуск

```bash
# Сборка
go build -o gocrawl ./cmd/gocrawl

# Запуск
./gocrawl -file test.txt -workers 5 -output ./test_downloads
```
