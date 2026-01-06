package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Storage  StorageConfig
	JWT      JWTConfig
	Security SecurityConfig
}

type ServerConfig struct {
	Port         string
	Host         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

type DatabaseConfig struct {
	URL      string
	MaxConns int
	MinConns int
}

type StorageConfig struct {
	BasePath         string
	TempPath         string
	PreviewPath      string
	OrganizeByDate   bool
	ChunkSize        int64
	BufferSize       int
	MaxFileSize      int64
}

type JWTConfig struct {
	Secret string
	Expiry time.Duration
}

type SecurityConfig struct {
	EnableMagicByteCheck  bool
	EnableMimeCheck       bool
	EnableMalwareScanner  bool
	EnableDocumentCheck   bool
	EnableAntivirusScan   bool
	ClamAVAddress         string
	EnableQuarantine      bool
	QuarantinePath        string
	TempScanPath          string
}

func Load() (*Config, error) {
	readTimeout, err := time.ParseDuration(getEnv("READ_TIMEOUT", "60s"))
	if err != nil {
		return nil, fmt.Errorf("invalid READ_TIMEOUT: %w", err)
	}

	writeTimeout, err := time.ParseDuration(getEnv("WRITE_TIMEOUT", "60s"))
	if err != nil {
		return nil, fmt.Errorf("invalid WRITE_TIMEOUT: %w", err)
	}

	jwtExpiry, err := time.ParseDuration(getEnv("JWT_EXPIRY", "24h"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_EXPIRY: %w", err)
	}

	cfg := &Config{
		Server: ServerConfig{
			Port:         getEnv("SERVER_PORT", "8080"),
			Host:         getEnv("SERVER_HOST", "0.0.0.0"),
			ReadTimeout:  readTimeout,
			WriteTimeout: writeTimeout,
		},
		Database: DatabaseConfig{
			URL:      getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/docvault?sslmode=disable"),
			MaxConns: getEnvAsInt("DB_MAX_CONNS", 25),
			MinConns: getEnvAsInt("DB_MIN_CONNS", 5),
		},
		Storage: StorageConfig{
			BasePath:       getEnv("STORAGE_BASE_PATH", "/data/documents"),
			TempPath:       getEnv("STORAGE_TEMP_PATH", "/data/documents/temp"),
			PreviewPath:    getEnv("STORAGE_PREVIEW_PATH", "/data/documents/preview"),
			OrganizeByDate: getEnvAsBool("STORAGE_ORGANIZE_BY_DATE", true),
			ChunkSize:      getEnvAsInt64("STORAGE_CHUNK_SIZE", 5242880),
			BufferSize:     getEnvAsInt("STORAGE_BUFFER_SIZE", 262144),
			MaxFileSize:    getEnvAsInt64("STORAGE_MAX_FILE_SIZE", 104857600),
		},
		JWT: JWTConfig{
			Secret: getEnv("JWT_SECRET", "your-super-secret-key-change-in-production"),
			Expiry: jwtExpiry,
		},
		Security: SecurityConfig{
			EnableMagicByteCheck: getEnvAsBool("ENABLE_MAGIC_BYTE_CHECK", true),
			EnableMimeCheck:      getEnvAsBool("ENABLE_MIME_CHECK", true),
			EnableMalwareScanner: getEnvAsBool("ENABLE_MALWARE_SCANNER", true),
			EnableDocumentCheck:  getEnvAsBool("ENABLE_DOCUMENT_CHECK", true),
			EnableAntivirusScan:  getEnvAsBool("ENABLE_ANTIVIRUS_SCAN", false),
			ClamAVAddress:        getEnv("CLAMAV_ADDRESS", "localhost:3310"),
			EnableQuarantine:     getEnvAsBool("ENABLE_QUARANTINE", true),
			QuarantinePath:       getEnv("QUARANTINE_PATH", "/data/quarantine"),
			TempScanPath:         getEnv("TEMP_SCAN_PATH", "/tmp/security-scan"),
		},
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}

func getEnvAsInt64(key string, defaultValue int64) int64 {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.ParseInt(valueStr, 10, 64)
	if err != nil {
		return defaultValue
	}
	return value
}

func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.ParseBool(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}
