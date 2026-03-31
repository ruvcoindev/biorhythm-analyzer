#!/bin/bash
# 🛡️ Git Safe AI Helper — Безопасная работа с AI
# Путь: ~/biorhythm-analyzer/scripts/git-safe-ai.sh

set -e
PROJECT_DIR="$HOME/biorhythm-analyzer"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

# Создание чекпоинта (ветка) — БЕЗ коммита, если код не компилируется
create_checkpoint() {
    BRANCH_NAME="checkpoint_$TIMESTAMP"
    echo "📍 Создание чекпоинта: $BRANCH_NAME"
    
    # Проверяем сборку
    if [ -f "go.mod" ]; then
        echo "🔍 Проверка сборки перед чекпоинтом..."
        if ! go build ./... >/dev/null 2>&1; then
            echo "⚠️  ВНИМАНИЕ: Код НЕ компилируется!"
            echo "Создаю ветку без коммита (исправь ошибки вручную)"
            git checkout -b "$BRANCH_NAME"
            echo "✅ Ветка создана: $BRANCH_NAME"
            return 0
        fi
    fi
    
    # Если всё ок — полноценный чекпоинт
    git checkout -b "$BRANCH_NAME"
    git add -A
    git commit -m "CHECKPOINT: Перед AI-изменениями [$TIMESTAMP]" --no-verify
    git checkout -
    echo "✅ Чекпоинт создан: $BRANCH_NAME"
}

# Быстрый откат
quick_rollback() {
    echo "⏮️  Откат к последнему стабильному коммиту..."
    echo ""
    git status
    echo ""
    echo "Действия:"
    echo "1) Отменить ВСЕ изменения (git reset --hard HEAD)"
    echo "2) Сохранить изменения в stash + откат"
    echo "3) Только показать изменения (git diff)"
    echo "4) Выйти"
    echo ""
    
    if [ -n "$1" ]; then
        choice="$1"
    else
        read -p "Выбор (1-4): " choice
    fi
    
    case $choice in
        1)
            echo "⚠️  ВНИМАНИЕ: Все несохранённые изменения будут ПОТЕРЯНЫ!"
            if [ -z "$2" ]; then
                read -p "Вы уверены? (y/n): " confirm
            else
                confirm="$2"
            fi
            if [ "$confirm" = "y" ]; then
                git reset --hard HEAD
                git clean -fd
                echo "✅ Откат выполнен"
            else
                echo "Отменено"
            fi
            ;;
        2)
            STASH_NAME="rollback_stash_$TIMESTAMP"
            git stash push -m "$STASH_NAME" --include-untracked
            echo "✅ Изменения сохранены в stash: $STASH_NAME"
            echo "   Для восстановления: git stash pop"
            ;;
        3)
            git diff
            ;;
        4|*)
            echo "Выход"
            ;;
    esac
}

# История изменений
show_history() {
    echo "📜 Последние 10 коммитов:"
    git log --oneline -10
    echo ""
    echo "📍 Чекпоинты (ветки):"
    git branch | grep "checkpoint_" | tail -10 || echo "  Нет чекпоинтов"
    echo ""
    echo "📦 Stash:"
    git stash list | head -5 || echo "  Пусто"
}

# Бэкап в stash
create_backup() {
    BACKUP_NAME="ai_backup_$TIMESTAMP"
    echo "📦 Создание бэкапа: $BACKUP_NAME"
    git stash push -m "$BACKUP_NAME" --include-untracked
    echo "✅ Бэкап создан. Для восстановления: git stash list && git stash pop"
}

# Восстановление из stash
restore_backup() {
    echo "🔍 Доступные бэкапы:"
    git stash list | grep -E "(ai_backup|rollback_stash)" || echo "  Нет бэкапов"
    echo ""
    if [ -n "$1" ]; then
        echo "♻️  Восстановление бэкапа #$1..."
        git stash pop "stash@{$1}" 2>/dev/null || git stash pop
        echo "✅ Восстановлено"
    else
        echo "💡 Укажите номер: $0 restore 0"
    fi
}

# Главная
case "${1:-help}" in
    checkpoint) create_checkpoint ;;
    rollback) quick_rollback "$2" "$3" ;;
    history) show_history ;;
    backup) create_backup ;;
    restore) restore_backup "$2" ;;
    help|*)
        echo "🛡️  Git Safe AI Helper — локальная версия"
        echo ""
        echo "⚡ Быстрый старт (перед запросом к AI):"
        echo "   ./scripts/git-safe-ai.sh checkpoint"
        echo ""
        echo "📋 Команды:"
        echo "   checkpoint  — точка восстановления"
        echo "   backup      — сохранить в stash"
        echo "   restore [N] — восстановить из stash"
        echo "   rollback    — интерактивный откат"
        echo "   history     — показать историю"
        echo ""
        echo "🚨 Аварийный откат:"
        echo "   ./scripts/git-safe-ai.sh rollback 1 y"
        echo ""
        echo "💡 Совет: Всегда делай checkpoint перед изменениями от AI!"
        ;;
esac
