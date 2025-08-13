package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Database DatabaseConfig
	Server   ServerConfig
	Files    FilesConfig
	Log      LogConfig
	JWT      JWTConfig
	CORS     CORSConfig
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	SSLMode  string
}

type ServerConfig struct {
	Host string
	Port int
}

type FilesConfig struct {
	ExcelOutputDir string
	TempDir        string
}

type LogConfig struct {
	Level  string
	Format string
}

type JWTConfig struct {
	Secret string
}

type CORSConfig struct {
	AllowedOrigins []string
}

func Load() (*Config, error) {
	// Cargar variables de entorno desde .env si existe
	if err := godotenv.Load(); err != nil {
		log.Println("No se encontró archivo .env, usando variables de entorno del sistema")
	}

	config := &Config{
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvAsInt("DB_PORT", 5432),
			User:     getEnv("DB_USER", "goexcel_user"),
			Password: getEnv("DB_PASSWORD", ""),
			Name:     getEnv("DB_NAME", "goexcel_db"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Server: ServerConfig{
			Host: getEnv("SERVER_HOST", "localhost"),
			Port: getEnvAsInt("SERVER_PORT", 8080),
		},
		Files: FilesConfig{
			ExcelOutputDir: getEnv("EXCEL_OUTPUT_DIR", "./output"),
			TempDir:        getEnv("TEMP_DIR", "./temp"),
		},
		Log: LogConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
		},
		JWT: JWTConfig{
			Secret: getEnv("JWT_SECRET", "default_secret_change_in_production"),
		},
		CORS: CORSConfig{
			AllowedOrigins: getEnvAsSlice("CORS_ALLOWED_ORIGINS", []string{"http://localhost:3000"}),
		},
	}

	return config, nil
}

func (c *Config) GetDatabaseURL() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host, c.Database.Port, c.Database.User,
		c.Database.Password, c.Database.Name, c.Database.SSLMode)
}

func (c *Config) GetServerAddress() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}

// Funciones helper para leer variables de entorno
func getEnv(key, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}

func getEnvAsInt(name string, defaultVal int) int {
	valueStr := getEnv(name, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultVal
}

func getEnvAsSlice(name string, defaultVal []string) []string {
	valueStr := getEnv(name, "")
	if valueStr == "" {
		return defaultVal
	}
	return strings.Split(valueStr, ",")
}