// Package entity ...
package entity

import (
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

// LogRecord Сущность "Запись в журнале"
type LogRecord struct {
	ID       uint64    `json:"id"`
	LogTime  time.Time `json:"logTime"`
	RealTime time.Time `json:"realTime"`
	Level    int       `json:"level"`
	Message1 string    `json:"message1"`
	Message2 string    `json:"message2"`
	Message3 string    `json:"message3"`
}

// IsEmpty ...
func (l *LogRecord) IsEmpty() bool {
	return l.ID == 0
}

// Validate ...
func (l *LogRecord) Validate() error {
	return validation.ValidateStruct(
		l,
		validation.Field(&l.LogTime, validation.Required),
		validation.Field(&l.Level, validation.Required),
		validation.Field(&l.Message1, validation.Required),
	)
}
