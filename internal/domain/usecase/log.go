// Package usecase Модели данных, относящиеся к журналу. Сейчас все в одном файле.
// При большом количестве моделей и операций имеет смысл разбить на несколько файлов или каталогов
package usecase

import (
	"time"

	"github.com/n-r-w/log-server-v2/internal/domain/entity"
)

type logUseCase struct {
	repo LogInterface
}

func NewLogCase(r LogInterface) *logUseCase {
	return &logUseCase{
		repo: r,
	}
}

func (l *logUseCase) Insert(logs []entity.LogRecord) error {
	return l.repo.Insert(logs)
}

func (l *logUseCase) Find(dateFrom time.Time, dateTo time.Time, limit int) (records []entity.LogRecord, limited bool, err error) {
	r, lim, e := l.repo.Find(dateFrom, dateTo, limit)
	return r, lim, e
}
