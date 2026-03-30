#!/bin/bash

set -e

echo "🔧 Настройка проекта Biorhythm Analyzer..."

# Проверка Go
if ! command -v go &> /dev/null; then
    echo "❌ Go не установлен. Установите Go 1.16+"
    exit 1
fi

# Инициализация модуля
go mod init biorhythm-analyzer 2>/dev/null || true

# Установка зависимостей
go get github.com/go-echarts/go-echarts/v2/...

# Создание директорий
mkdir -p data static/templates

# Инициализация файлов данных
if [ ! -f data/people.json ]; then
    echo '[]' > data/people.json
    echo "✅ Создан data/people.json"
fi

if [ ! -f data/history.log.json ]; then
    echo '[]' > data/history.log.json
    echo "✅ Создан data/history.log.json"
fi

echo "✅ Настройка завершена!"
echo ""
echo "📝 Команды для работы:"
echo "   make build    - Собрать проект"
echo "   make run      - Запустить анализ"
echo "   make advanced - Расширенный анализ"
echo "   make add      - Добавить человека"
echo "   make list     - Показать всех"
