package main

import (
	"fmt"
	"log"
	"os"
	"qweasley/internal/database"
	"strings"
	"time"
)

func main() {
	fmt.Println("🔍 Проверка соединения с базой данных...")

	// Загружаем переменные окружения
	if os.Getenv("LOCAL_TEST") == "true" {
		loadEnvFile()
	}

	// Проверяем наличие обязательных переменных окружения
	requiredEnvVars := []string{"DB_HOST", "DB_USER", "DB_PASSWORD", "DB_NAME"}
	for _, envVar := range requiredEnvVars {
		if os.Getenv(envVar) == "" {
			log.Fatalf("❌ Переменная окружения %s не установлена", envVar)
		}
	}

	fmt.Printf("📊 Параметры подключения:\n")
	fmt.Printf("   Host: %s\n", os.Getenv("DB_HOST"))
	fmt.Printf("   Port: %s\n", getEnvWithDefault("DB_PORT", "5432"))
	fmt.Printf("   Database: %s\n", os.Getenv("DB_NAME"))
	fmt.Printf("   User: %s\n", os.Getenv("DB_USER"))

	// Инициализируем базу данных
	fmt.Println("\n🔌 Подключение к базе данных...")
	startTime := time.Now()

	if err := database.InitDatabase(); err != nil {
		log.Fatalf("❌ Ошибка подключения к базе данных: %v", err)
	}

	connectionTime := time.Since(startTime)
	fmt.Printf("✅ Подключение успешно установлено за %v\n", connectionTime)

	db := database.GetDB()

	// Проверяем доступность таблиц (только чтение)
	fmt.Println("\n📋 Проверка структуры базы данных...")

	tables := []string{"chats", "questions", "pictures", "feedbacks", "reactions"}
	for _, table := range tables {
		var count int
		if err := db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", table)).Scan(&count); err != nil {
			fmt.Printf("⚠️  Таблица %s недоступна: %v\n", table, err)
		} else {
			fmt.Printf("✅ Таблица %s доступна\n", table)
		}
	}

	// Проверяем количество записей в основных таблицах
	fmt.Println("\n📊 Статистика базы данных:")

	var count int

	// Количество вопросов
	if err := db.QueryRow("SELECT COUNT(*) FROM questions").Scan(&count); err != nil {
		fmt.Printf("⚠️  Не удалось получить количество вопросов: %v\n", err)
	} else {
		fmt.Printf("   Вопросов: %d\n", count)
	}

	// Количество опубликованных вопросов
	if err := db.QueryRow("SELECT COUNT(*) FROM questions WHERE is_published = true").Scan(&count); err != nil {
		fmt.Printf("⚠️  Не удалось получить количество опубликованных вопросов: %v\n", err)
	} else {
		fmt.Printf("   Опубликованных вопросов: %d\n", count)
	}

	// Количество чатов
	if err := db.QueryRow("SELECT COUNT(*) FROM chats").Scan(&count); err != nil {
		fmt.Printf("⚠️  Не удалось получить количество чатов: %v\n", err)
	} else {
		fmt.Printf("   Чатов: %d\n", count)
	}

	// Количество обращений
	if err := db.QueryRow("SELECT COUNT(*) FROM feedbacks").Scan(&count); err != nil {
		fmt.Printf("⚠️  Не удалось получить количество обращений: %v\n", err)
	} else {
		fmt.Printf("   Обращений: %d\n", count)
	}

	// Тестируем простой запрос
	fmt.Println("\n🧪 Тестирование запросов...")

	// Пробуем получить случайный вопрос
	var questionText string
	if err := db.QueryRow("SELECT text FROM questions WHERE is_published = true LIMIT 1").Scan(&questionText); err != nil {
		fmt.Printf("⚠️  Не удалось получить тестовый вопрос: %v\n", err)
	} else {
		fmt.Printf("✅ Тестовый запрос выполнен успешно\n")
		fmt.Printf("   Получен вопрос: %s\n", truncateString(questionText, 50))
	}

	// Проверяем производительность
	fmt.Println("\n⚡ Тест производительности...")

	startTime = time.Now()
	for i := 0; i < 5; i++ {
		var result int
		if err := db.QueryRow("SELECT 1").Scan(&result); err != nil {
			log.Fatalf("❌ Ошибка при тестировании производительности: %v", err)
		}
	}
	avgQueryTime := time.Since(startTime) / 5
	fmt.Printf("✅ Среднее время запроса: %v\n", avgQueryTime)

	fmt.Println("\n🎉 Проверка соединения с базой данных завершена успешно!")
}

func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func loadEnvFile() {
	file, err := os.Open(".env")
	if err != nil {
		return
	}
	defer file.Close()

	content, err := os.ReadFile(".env")
	if err != nil {
		return
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			os.Setenv(key, value)
		}
	}
}
