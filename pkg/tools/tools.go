// Package tools Различные фукции общего назначения
package tools

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"fmt"
	"io"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

// EncryptPassword Генерация хэша пароля
func EncryptPassword(s string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(strings.TrimSpace(s)), bcrypt.MinCost)
	if err != nil {
		return "", fmt.Errorf("failed GenerateFromPassword %v ", err)
	}

	return string(b), nil
}

// ComparePassword Подходит ли пароль
func ComparePassword(encryptedPassword string, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(encryptedPassword), []byte(password)) == nil
}

// RequiredIf Валидатор для проверки по условию
func RequiredIf(cond bool) validation.RuleFunc {
	return func(value interface{}) error {
		if cond {
			return fmt.Errorf("failed validation %v", validation.Validate(value, validation.Required))
		}

		return nil
	}
}

// CompressData Сжатие массива данных
func CompressData(deflateCompression bool, data []byte) (resData []byte, err error) {
	if data == nil {
		return []byte{}, nil
	}

	// алгоритм сжатия
	var compressor io.WriteCloser
	// целевой буфер
	var compressedBuf bytes.Buffer

	// сжимаем по нужному алгоритму
	if deflateCompression {
		if compressor, err = flate.NewWriter(&compressedBuf, flate.BestSpeed); err != nil {
			return nil, errors.Wrap(err, "deflate error")
		}
	} else {
		if compressor, err = gzip.NewWriterLevel(&compressedBuf, gzip.BestSpeed); err != nil {
			return nil, errors.Wrap(err, "gzip error")
		}
	}

	if _, err := compressor.Write(data); err != nil {
		return nil, errors.Wrap(err, "compress error")
	}

	if err := compressor.Close(); err != nil {
		return nil, errors.Wrap(err, "compress error")
	}

	return compressedBuf.Bytes(), nil
}
