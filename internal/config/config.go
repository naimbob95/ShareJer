package config

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all runtime configuration, sourced from environment variables
// (optionally seeded from a .env file). Every field has a sensible default so
// the app runs with zero configuration.
type Config struct {
	Port          string        // HTTP listen port, e.g. "8080"
	BaseURL       string        // public base URL used in share links + QR codes
	MaxUploadSize int64         // max upload size in bytes
	FileExpiry    time.Duration // how long after upload a file is reachable (0 = never expires)
	CleanupEvery  time.Duration // how often the background sweeper removes expired files
	DBPath        string        // SQLite file path
	UploadsDir    string        // directory for stored files
}

// Load reads configuration from the environment. If a .env file exists in the
// working directory, its values are loaded first (without overriding variables
// already set in the real environment).
func Load() Config {
	loadDotEnv(".env")

	return Config{
		Port:          getEnv("PORT", "8080"),
		BaseURL:       getEnv("BASE_URL", "http://localhost:8080"),
		MaxUploadSize: getEnvInt("MAX_UPLOAD_SIZE_MB", 5) << 20,
		FileExpiry:    getEnvDuration("FILE_EXPIRY", 15*time.Minute),
		CleanupEvery:  getEnvDuration("CLEANUP_INTERVAL", 5*time.Minute),
		DBPath:        getEnv("DB_PATH", "sharejer.db"),
		UploadsDir:    getEnv("UPLOADS_DIR", "uploads"),
	}
}

// loadDotEnv reads simple KEY=VALUE lines from path into the environment.
// Missing file is not an error. Existing environment variables win.
func loadDotEnv(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, found := strings.Cut(line, "=")
		if !found {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.Trim(strings.TrimSpace(value), `"'`)
		if _, exists := os.LookupEnv(key); !exists {
			os.Setenv(key, value)
		}
	}
	_ = scanner.Err() // best-effort: a partially read .env still applies what it parsed
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int64) int64 {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			return n
		}
	}
	return fallback
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
}
