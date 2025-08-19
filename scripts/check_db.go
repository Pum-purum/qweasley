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
		if err := db.Raw(fmt.Sprintf("SELECT COUNT(*) FROM %s", table)).Scan(&count).Error; err != nil {
			fmt.Printf("⚠️ Таблица %s недоступна: %v\n", table, err)
		} else {
			fmt.Printf("✅ Таблица %s доступна\n", table)
		}
	}

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
