# 🎯 Qweasley - Telegram Quiz Bot

[![Go Version](https://img.shields.io/badge/Go-1.23+-blue.svg)](https://golang.org/)
[![Telegram Bot API](https://img.shields.io/badge/Telegram%20Bot%20API-v5.5.1-green.svg)](https://core.telegram.org/bots/api)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-Supported-blue.svg)](https://www.postgresql.org/)
[![Docker](https://img.shields.io/badge/Docker-Supported-blue.svg)](https://www.docker.com/)

> Интеллектуальный Telegram-бот для проведения увлекательных квизов с системой монет и обратной связи.
> Бот задает вопрос - вы отвечаете. Все просто. Изначально бот был написан на PHP, впоследствии переписан на Go.
> PHP-версию вы можете посмотреть в ветке php.

## 🏗️ Архитектура проекта

```
qweasley_go/
├── cmd/function/          # Точка входа приложения
├── internal/
│   ├── database/         # Конфигурация БД
│   ├── handlers/         # Обработчики команд и сообщений
│   ├── models/           # Модели данных
│   └── repository/       # Слой доступа к данным
├── crt/                  # Сертификаты
├── deploy.sh            # Скрипт развертывания
└── docker-compose.yml   # Конфигурация Docker
```

## 🛠️ Технологический стек

- **Backend**: Go 1.23+
- **Telegram API**: go-telegram-bot-api v5.5.1
- **Database**: Облачная PostgreSQL в NeonTech
- **Deployment**: Yandex Cloud Functions
- **Containerization**: Docker & Docker Compose
- **Architecture**: Clean Architecture с разделением слоев

## 🚀 Быстрый старт

### Установка и запуск

1. **Клонирование репозитория**
```bash
git clone <repository-url>
cd qweasley_go
```

2. **Настройка переменных окружения**
```bash
cp .env.example .env
# Отредактируйте .env файл с вашими настройками
```

3. **Запуск с Docker**
```bash
make up          # Запуск контейнеров
make dev         # Запуск в режиме разработки
```

## 🚀 Развертывание

### Yandex Cloud Functions
```bash
make deploy    # Автоматическое развертывание
```

### Локальное тестирование
```bash
make dev       # Запуск в режиме разработки
make check     # Проверка подключения к БД
```

## 📄 Лицензия

Этот проект распространяется под лицензией MIT. См. файл `LICENSE` для подробностей.

## 📞 Поддержка

Если у вас есть вопросы или предложения, создайте issue в репозитории или свяжитесь с командой разработки.

---

**Qweasley** - где каждый вопрос становится приключением! 🎯✨
