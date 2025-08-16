#!/bin/bash

# Простой скрипт развертывания Telegram бота в Яндекс.Облако

set -e

# Загрузка переменных из .env файла
if [ -f ".env" ]; then
    source .env
fi

# Конфигурация
FUNCTION_NAME="goqweasley"
MEMORY="128m"
TIMEOUT="2s"
RUNTIME="golang121"

echo "🚀 Развертывание бота в Яндекс.Облако..."

# Проверки
if ! command -v yc &> /dev/null; then
    echo "❌ Установите Yandex Cloud CLI: curl -sSL https://storage.yandexcloud.net/yandexcloud-yc/install.sh | bash"
    exit 1
fi

if [ -z "$FOLDER_ID" ]; then
    echo "❌ Установите FOLDER_ID в .env файле"
    exit 1
fi

if [ -z "$SERVICE_ACCOUNT_ID" ]; then
    echo "❌ Установите SERVICE_ACCOUNT_ID в .env файле"
    exit 1
fi

# Подготовка кода
echo "📦 Подготовка кода..."
BUILD_DIR="./build"
rm -rf $BUILD_DIR
mkdir -p $BUILD_DIR

# Очистка от старых бинарников
rm -f main function.zip

# Проверка необходимых файлов
if [ ! -f "cmd/function/main.go" ]; then
    echo "❌ Файл cmd/function/main.go не найден"
    exit 1
fi

if [ ! -f "go.mod" ]; then
    echo "❌ Файл go.mod не найден"
    exit 1
fi

# Копируем все файлы, сохраняя структуру модуля
cp go.mod $BUILD_DIR/
cp go.sum $BUILD_DIR/ 2>/dev/null || true
cp cmd/function/main.go $BUILD_DIR/

# Создание архива с исходным кодом
echo "🔧 Создание архива..."
cd $BUILD_DIR

# Создаем архив только с исходным кодом (исключаем бинарники)
zip -r function.zip . -x "*.exe" "main" "*.so" "*.dylib" > /dev/null

cd ..

# Развертывание
echo "☁️ Развертывание..."
yc serverless function version create \
    --function-name=$FUNCTION_NAME \
    --folder-id=$FOLDER_ID \
    --runtime=$RUNTIME \
    --entrypoint=main.Handler \
    --memory=$MEMORY \
    --execution-timeout=$TIMEOUT \
    --source-path=$BUILD_DIR/function.zip \
    --service-account-id=$SERVICE_ACCOUNT_ID \
    --environment TELEGRAM_TOKEN="$TELEGRAM_TOKEN"

# Получение URL и настройка webhook
FUNCTION_ID=$(yc serverless function get $FUNCTION_NAME --folder-id=$FOLDER_ID --format=json | jq -r '.id')
INVOKE_URL="https://functions.yandexcloud.net/$FUNCTION_ID"

echo "✅ Функция развернута: $INVOKE_URL"

# Очистка старых версий (оставляем только последние 3)
echo "🧹 Очистка старых версий..."
OLD_VERSIONS=$(yc serverless function version list --function-name=$FUNCTION_NAME --folder-id=$FOLDER_ID --format=json | jq -r '.[3:] | .[].id')
if [ ! -z "$OLD_VERSIONS" ]; then
    for version_id in $OLD_VERSIONS; do
        echo "Удаление версии: $version_id"
        yc serverless function version delete --id=$version_id
    done
fi

# Очистка
rm -rf $BUILD_DIR

echo "🎉 Готово! Отправьте /start боту для проверки."