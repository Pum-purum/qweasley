package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// Config конфигурация базы данных
type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
	CertPath string
}

// LoadConfig загружает конфигурацию из переменных окружения
func LoadConfig() *Config {
	config := &Config{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnv("DB_PORT", "5432"),
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		Name:     os.Getenv("DB_NAME"),
		SSLMode:  getEnv("SSL_MODE", "require"),
		CertPath: os.Getenv("SSL_CERT_PATH"),
	}

	// Проверяем обязательные параметры
	if config.User == "" || config.Password == "" || config.Name == "" {
		log.Fatal("DB_USER, DB_PASSWORD, and DB_NAME environment variables are required")
	}

	return config
}

// getEnv получает значение переменной окружения или возвращает значение по умолчанию
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// InitDatabase инициализирует подключение к базе данных
func InitDatabase() error {
	config := LoadConfig()

	// Формируем строку подключения
	// Для Neon.tech используем sslmode=prefer для обхода проблем с сертификатами
	dsn := fmt.Sprintf("host=%s port=%s dbname=%s sslmode=prefer user=%s password=%s",
		config.Host, config.Port, config.Name, config.User, config.Password)

	// Настройки GORM
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	}

	// Подключаемся к базе данных
	gormDB, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return fmt.Errorf("failed to open database connection: %w", err)
	}

	// Получаем sql.DB для настройки пула соединений
	sqlDB, err := gormDB.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}

	// Настраиваем пул соединений
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Проверяем соединение
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	// Сохраняем подключение
	DB = gormDB

	log.Println("Database connection established successfully")
	return nil
}

// GetDB возвращает экземпляр базы данных
func GetDB() *gorm.DB {
	if DB == nil {
		log.Fatal("Database connection not initialized. Call InitDatabase() first.")
	}
	return DB
}
