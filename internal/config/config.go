package config

import (
	"github.com/BurntSushi/toml"
)

// Конфиг logserver.toml
type Config struct {
	SuperAdminID            uint64
	Host                    string `toml:"HOST"`
	Port                    string `toml:"PORT"`
	SuperAdminLogin         string `toml:"SUPERADMIN_LOGIN"`
	SuperPassword           string `toml:"SUPERADMIN_PASSWORD"`
	SessionAge              uint   `toml:"SESSION_AGE"`
	LogLevel                string `toml:"LOG_LEVEL"`
	DatabaseURL             string `toml:"DATABASE_URL"`
	SessionEncriptionKey    string `toml:"SESSION_ENCRYPTION_KEY"`
	MaxDbSessions           uint   `toml:"MAX_DB_SESSIONS"`
	MaxDbSessionIdleTimeSec uint   `toml:"MAX_DB_SESSION_IDLE_TIME_SEC"`
	MaxLogRecordsResult     uint   `toml:"MAX_LOG_RECORDS_RESULT"`
	MaxLogRecordsResultWeb  uint   `toml:"MAX_LOG_RECORDS_RESULT_WEB"`
	PasswordRegex           string `toml:"PASSWORD_REGEX"`
	PasswordRegexError      string `toml:"PASSWORD_REGEX_ERROR"`
	HttpReadTimeout         uint   `toml:"HTTP_READ_TIMEOUT"`
	HttpWriteTimeout        uint   `toml:"HTTP_WRITE_TIMEOUT"`
	HttpShutdownTimeout     uint   `toml:"HTTP_SHUTDOWN_TIMEOUT"`
}

const (
	superAdminID            = 1
	maxDbSessions           = 50
	maxDbSessionIdleTimeSec = 50
	maxLogRecordsResult     = 100000
	maxLogRecordsResultWeb  = 1000
	defaultSessionAge       = 60 * 60 * 24 // 24 часа
)

// Load Инициализация конфига значениями по умолчанию
func New(path string) (*Config, error) {
	c := &Config{
		SuperAdminID:            superAdminID,
		Host:                    "0.0.0.0",
		Port:                    "8080",
		SuperAdminLogin:         "admin",
		SuperPassword:           "admin",
		SessionAge:              defaultSessionAge,
		LogLevel:                "debug",
		DatabaseURL:             "log",
		SessionEncriptionKey:    "e09469b1507d0e7a98831750aff903e0831a428f9addf3cfa348fa64dcf",
		MaxDbSessions:           maxDbSessions,
		MaxDbSessionIdleTimeSec: maxDbSessionIdleTimeSec,
		MaxLogRecordsResult:     maxLogRecordsResult,
		MaxLogRecordsResultWeb:  maxLogRecordsResultWeb,
		// PasswordRegex:           "^[A-Za-z0-9@$!%*?&]{8,}$",
		PasswordRegex:       ".*",
		PasswordRegexError:  "Латинские буквы, цифры и символы @$!%*?& без пробелов, минимум 4 символа",
		HttpReadTimeout:     5,
		HttpWriteTimeout:    5,
		HttpShutdownTimeout: 10,
	}

	if path == "" {
		return c, nil
	}

	if _, err := toml.DecodeFile(path, c); err != nil {
		return nil, err
	}

	return c, nil
}
