// Package model Модели данных, относящиеся к журналу.
// Сейчас все в одном файле. При большом количестве моделей и операций
// имеет смысл разбить на несколько файлов или каталогов
package entity

import (
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type LogRecord struct {
	ID       uint64    `json:"id"`
	LogTime  time.Time `json:"logTime"`
	RealTime time.Time `json:"realTime"`
	Level    uint      `json:"level"`
	Message1 string    `json:"message1"`
	Message2 string    `json:"message2"`
	Message3 string    `json:"message3"`
}

func (u *LogRecord) IsEmpty() bool {
	return u.ID == 0
}

func (l *LogRecord) Validate() error {
	return validation.ValidateStruct(
		l,
		validation.Field(&l.LogTime, validation.Required),
		validation.Field(&l.Level, validation.Required),
		validation.Field(&l.Message1, validation.Required),
	)
}
