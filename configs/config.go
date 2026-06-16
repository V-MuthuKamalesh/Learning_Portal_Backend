package configs

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all runtime configuration, loaded from environment variables.
type Config struct {
	App     AppConfig
	DB      DBConfig
	Redis   RedisConfig
	JWT     JWTConfig
	Auth    AuthConfig
	Storage StorageConfig
	Judge   JudgeConfig
	SMTP    SMTPConfig
	Cors    []string
	Seed    SeedConfig
}

type AppConfig struct {
	Env  string
	Port string
	Name string
}

type DBConfig struct {
	Host, Port, User, Password, Name, SSLMode string
}

type RedisConfig struct {
	Enabled  bool
	Host     string
	Port     string
	Password string
	DB       int
}

type JWTConfig struct {
	AccessSecret  string
	RefreshSecret string
	AccessTTL     time.Duration
	RefreshTTL    time.Duration
}

type AuthConfig struct {
	MaxFailedAttempts int
	LockoutDuration   time.Duration
}

type StorageConfig struct {
	Driver    string
	LocalDir  string
	PublicURL string
	S3Bucket  string
	S3Region  string
	S3Access  string
	S3Secret  string
	S3Endpoint string
}

type JudgeConfig struct {
	URL     string
	Enabled bool
}

type SMTPConfig struct {
	Host, Port, User, Password, From string
}

type SeedConfig struct {
	SuperAdminEmail    string
	SuperAdminPassword string
	CollegeName        string
	CollegeCode        string
}

// Load reads configuration from the environment, applying sensible defaults.
func Load() *Config {
	return &Config{
		App: AppConfig{
			Env:  env("APP_ENV", "development"),
			Port: env("APP_PORT", "8080"),
			Name: env("APP_NAME", "college-assessment-api"),
		},
		DB: DBConfig{
			Host:     env("DB_HOST", "localhost"),
			Port:     env("DB_PORT", "5432"),
			User:     env("DB_USER", "postgres"),
			Password: env("DB_PASSWORD", "postgres"),
			Name:     env("DB_NAME", "college_assess"),
			SSLMode:  env("DB_SSLMODE", "disable"),
		},
		Redis: RedisConfig{
			Enabled:  envBool("REDIS_ENABLED", false),
			Host:     env("REDIS_HOST", "localhost"),
			Port:     env("REDIS_PORT", "6379"),
			Password: env("REDIS_PASSWORD", ""),
			DB:       envInt("REDIS_DB", 0),
		},
		JWT: JWTConfig{
			AccessSecret:  env("JWT_ACCESS_SECRET", "dev-access-secret"),
			RefreshSecret: env("JWT_REFRESH_SECRET", "dev-refresh-secret"),
			AccessTTL:     time.Duration(envInt("JWT_ACCESS_TTL_MINUTES", 15)) * time.Minute,
			RefreshTTL:    time.Duration(envInt("JWT_REFRESH_TTL_HOURS", 720)) * time.Hour,
		},
		Auth: AuthConfig{
			MaxFailedAttempts: envInt("MAX_FAILED_ATTEMPTS", 5),
			LockoutDuration:   time.Duration(envInt("LOCKOUT_MINUTES", 15)) * time.Minute,
		},
		Storage: StorageConfig{
			Driver:     env("STORAGE_DRIVER", "local"),
			LocalDir:   env("STORAGE_LOCAL_DIR", "./uploads"),
			PublicURL:  env("STORAGE_PUBLIC_URL", "http://localhost:8080/uploads"),
			S3Bucket:   env("S3_BUCKET", ""),
			S3Region:   env("S3_REGION", ""),
			S3Access:   env("S3_ACCESS_KEY", ""),
			S3Secret:   env("S3_SECRET_KEY", ""),
			S3Endpoint: env("S3_ENDPOINT", ""),
		},
		Judge: JudgeConfig{
			URL:     env("JUDGE_URL", "http://localhost:9090"),
			Enabled: envBool("JUDGE_ENABLED", true),
		},
		SMTP: SMTPConfig{
			Host:     env("SMTP_HOST", "localhost"),
			Port:     env("SMTP_PORT", "1025"),
			User:     env("SMTP_USER", ""),
			Password: env("SMTP_PASSWORD", ""),
			From:     env("SMTP_FROM", "no-reply@college-assess.local"),
		},
		Cors: strings.Split(env("CORS_ORIGINS", "http://localhost:3000,http://localhost:3001"), ","),
		Seed: SeedConfig{
			SuperAdminEmail:    env("SEED_SUPERADMIN_EMAIL", "superadmin@demo.edu"),
			SuperAdminPassword: env("SEED_SUPERADMIN_PASSWORD", "Admin@12345"),
			CollegeName:        env("SEED_COLLEGE_NAME", "Demo College of Engineering"),
			CollegeCode:        env("SEED_COLLEGE_CODE", "DEMO"),
		},
	}
}

func (c *Config) IsProd() bool { return c.App.Env == "production" }

func env(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return def
}

func envInt(key string, def int) int {
	if v, ok := os.LookupEnv(key); ok {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func envBool(key string, def bool) bool {
	if v, ok := os.LookupEnv(key); ok {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return def
}
