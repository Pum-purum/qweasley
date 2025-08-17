package database

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"log"
	"os"
)

var DB *sql.DB

// InitDatabase инициализирует подключение к базе данных
func InitDatabase() error {
	// Получаем параметры подключения из переменных окружения
	dbHost := os.Getenv("DB_HOST")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbPort := os.Getenv("DB_PORT")

	// Устанавливаем значения по умолчанию, если переменные не заданы
	if dbPort == "" {
		dbPort = "5432"
	}
	if dbHost == "" {
		dbHost = "localhost"
	}

	// Проверяем обязательные параметры
	if dbUser == "" || dbPassword == "" || dbName == "" {
		return fmt.Errorf("DB_USER, DB_PASSWORD, and DB_NAME environment variables are required")
	}

	// Формируем строку подключения как в PHP
	dsn := fmt.Sprintf("host=%s port=%s dbname=%s sslmode=require user=%s password=%s",
		dbHost, dbPort, dbName, dbUser, dbPassword)

	log.Printf("Connecting to database: host=%s, port=%s, dbname=%s, user=%s",
		dbHost, dbPort, dbName, dbUser)

	// Подключаемся через database/sql
	sqlDB, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("failed to open database connection: %w", err)
	}

	// Проверяем подключение
	if err := sqlDB.Ping(); err != nil {
		sqlDB.Close()
		return fmt.Errorf("failed to ping database: %w", err)
	}

	// Сохраняем подключение
	DB = sqlDB

	log.Println("Database connection established successfully")
	return nil
}

// GetDB возвращает экземпляр базы данных
func GetDB() *sql.DB {
	return DB
}

// CloseDatabase закрывает соединение с базой данных
func CloseDatabase() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}
