// Package config ...
package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/BurntSushi/toml"
	"github.com/n-r-w/log-server-v2/pkg/logger"
)

// Config logserver.toml
type Config struct {
	SuperAdminID            int
	Host                    string `toml:"HOST"`
	Port                    string `toml:"PORT"`
	SuperAdminLogin         string `toml:"SUPERADMIN_LOGIN"`
	SuperPassword           string `toml:"SUPERADMIN_PASSWORD"`
	SessionAge              int    `toml:"SESSION_AGE"`
	LogLevel                string `toml:"LOG_LEVEL"`
	DatabaseURL             string `toml:"DATABASE_URL"`
	SessionEncriptionKey    string `toml:"SESSION_ENCRYPTION_KEY"`
	MaxDbSessions           int    `toml:"MAX_DB_SESSIONS"`
	MaxDbSessionIdleTimeSec int    `toml:"MAX_DB_SESSION_IDLE_TIME_SEC"`
	MaxLogRecordsResult     int    `toml:"MAX_LOG_RECORDS_RESULT"`
	PasswordRegex           string `toml:"PASSWORD_REGEX"`
	PasswordRegexError      string `toml:"PASSWORD_REGEX_ERROR"`
	HttpReadTimeout         int    `toml:"HTTP_READ_TIMEOUT"`
	HttpWriteTimeout        int    `toml:"HTTP_WRITE_TIMEOUT"`
	HttpShutdownTimeout     int    `toml:"HTTP_SHUTDOWN_TIMEOUT"`
	RateLimit               int    `toml:"RATE_LIMIT"`
	RateLimitBurst          int    `toml:"RATE_LIMIT_BURST"`
}

const (
	superAdminID            = 1
	maxDbSessions           = 50
	maxDbSessionIdleTimeSec = 50
	maxLogRecordsResult     = 100000
	defaultSessionAge       = 60 * 60 * 24 // 24 часа
)

// New Инициализация конфига значениями по умолчанию
func New(path string, logger logger.Interface) (*Config, error) {
	c := &Config{
		SuperAdminID:            superAdminID,
		Host:                    "0.0.0.0",
		Port:                    "8080",
		SuperAdminLogin:         "admin",
		SuperPassword:           "admin",
		SessionAge:              defaultSessionAge,
		LogLevel:                "debug",
		DatabaseURL:             "",
		SessionEncriptionKey:    "e09469b1507d0e7a98831750aff903e0831a428f9addf3cfa348fa64dcf",
		MaxDbSessions:           maxDbSessions,
		MaxDbSessionIdleTimeSec: maxDbSessionIdleTimeSec,
		MaxLogRecordsResult:     maxLogRecordsResult,
		PasswordRegex:           ".*",
		PasswordRegexError:      "Латинские буквы, цифры и символы @$!%*?& без пробелов, минимум 4 символа",
		HttpReadTimeout:         5,
		HttpWriteTimeout:        5,
		HttpShutdownTimeout:     10,
		RateLimit:               10000,
		RateLimitBurst:          20000,
	}

	c.readEnv()

	if path != "" {
		if _, err := toml.DecodeFile(path, c); err != nil {
			return nil, err
		}
	}

	if c.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL undefined")
	}

	logger.Info("MAX_DB_SESSIONS: %d", c.MaxDbSessions)
	logger.Info("SESSION_AGE: %d", c.SessionAge)
	logger.Info("MAX_DB_SESSION_IDLE_TIME_SEC: %d", c.MaxDbSessionIdleTimeSec)
	logger.Info("RATE_LIMIT: %d", c.RateLimit)
	logger.Info("RATE_LIMIT_BURST: %d", c.RateLimitBurst)
	logger.Info("LS_DATABASE_URL: %s", c.DatabaseURL)

	return c, nil
}

// Чтение переменных окружения
func (c *Config) readEnv() {
	eString(&c.Host, "LS_HOST")
	eString(&c.Port, "LS_PORT")
	eString(&c.SuperAdminLogin, "LS_SUPERADMIN_LOGIN")
	eString(&c.SuperPassword, "LS_SUPERADMIN_PASSWORD")
	eInt(&c.SessionAge, "LS_SESSION_AGE")
	eString(&c.LogLevel, "LS_LOG_LEVEL")
	eString(&c.DatabaseURL, "LS_DATABASE_URL")
	eInt(&c.MaxDbSessions, "LS_MAX_DB_SESSIONS")
	eInt(&c.MaxDbSessionIdleTimeSec, "LS_MAX_DB_SESSION_IDLE_TIME_SEC")
	eString(&c.SessionEncriptionKey, "LS_SESSION_ENCRYPTION_KEY")
	eInt(&c.MaxLogRecordsResult, "LS_MAX_LOG_RECORDS_RESULT")
	eInt(&c.HttpReadTimeout, "LS_HTTP_READ_TIMEOUT")
	eInt(&c.HttpWriteTimeout, "LS_HTTP_WRITE_TIMEOUT")
	eInt(&c.HttpShutdownTimeout, "LS_HTTP_SHUTDOWN_TIMEOUT")
	eInt(&c.RateLimit, "LS_RATE_LIMIT")
	eInt(&c.RateLimitBurst, "LS_RATE_LIMIT_BURST")
	eString(&c.PasswordRegex, "LS_PASSWORD_REGEX")
	eString(&c.PasswordRegexError, "LS_PASSWORD_REGEX_ERROR")
}

func eString(dest *string, env string) {
	if e := os.Getenv(env); len(e) > 0 {
		*dest = e
	}
}

func eInt(dest *int, env string) {
	if e := os.Getenv(env); len(e) > 0 {
		if i, err := strconv.Atoi(e); err == nil {
			*dest = i
		}
	}
}
