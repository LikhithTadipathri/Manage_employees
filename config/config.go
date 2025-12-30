package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
	Server      ServerConfig
	Database    DatabaseConfig
	JWT         JWTConfig
	Logger      LoggerConfig
	TLS         TLSConfig
	Environment string
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port int
	Host string
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	Secret           string
	ExpirationHours  int
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host              string
	Port              int
	User              string
	Password          string
	DBName            string
	SSLMode           string
	MaxOpenConns      int
	MaxIdleConns      int
	ConnMaxLifetime   int
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	
	envPaths := []string{
		".env",                                    // Current directory
		filepath.Join("..", "..", ".env"),        // From cmd/employee-service
		filepath.Join("../.env"),                 // From subdirectories
	}

	for _, envPath := range envPaths {
		if _, err := os.Stat(envPath); err == nil {
			_ = godotenv.Load(envPath)
			break
		}
	}

	// If no .env found, try loading from current directory anyway
	_ = godotenv.Load()

	environment := getEnv("ENVIRONMENT", "development")
	loggerCfg := GetLoggerConfig(environment)

	config := &Config{
		Environment: environment,
		Server: ServerConfig{
			Port: getEnvAsInt("SERVER_PORT", 8080),
			Host: getEnv("SERVER_HOST", "0.0.0.0"),
		},
		Database: DatabaseConfig{
			Host:              getEnv("DB_HOST", "localhost"),
			Port:              getEnvAsInt("DB_PORT", 5432),
			User:              getEnv("DB_USER", "postgres"),
			Password:          getEnv("DB_PASSWORD", "postgres"),
			DBName:            getEnv("DB_NAME", "employee_db"),
			SSLMode:           getEnv("DB_SSLMODE", "disable"),
			MaxOpenConns:      getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:      getEnvAsInt("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime:   getEnvAsInt("DB_CONN_MAX_LIFETIME_MINUTES", 5),
		},
		JWT: JWTConfig{
			Secret:          getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
			ExpirationHours: getEnvAsInt("JWT_EXPIRATION_HOURS", 1),
		},
		Logger: LoggerConfig{
			Level:  getEnv("LOG_LEVEL", loggerCfg.Level),
			Format: getEnv("LOG_FORMAT", loggerCfg.Format),
			Output: getEnv("LOG_OUTPUT", loggerCfg.Output),
		},
		TLS: *LoadTLSConfig(),
	}

	return config, nil
}

// GetDatabaseURL returns the database connection string
func (c *Config) GetDatabaseURL() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host,
		c.Database.Port,
		c.Database.User,
		c.Database.Password,
		c.Database.DBName,
		c.Database.SSLMode,
	)
}

// Helper functions
func getEnv(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}

func getEnvAsInt(key string, defaultVal int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultVal
}
