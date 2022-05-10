// Package usecase Модели данных, относящиеся к журналу. Сейчас все в одном файле.
// При большом количестве моделей и операций имеет смысл разбить на несколько файлов или каталогов
package usecase

import (
	"time"

	"github.com/n-r-w/log-server-v2/internal/entity"
)

type LogUseCase struct {
	repo LogInterface
}

func NewLogCase(r LogInterface) *LogUseCase {
	return &LogUseCase{
		repo: r,
	}
}

func (l *LogUseCase) Insert(logs []entity.LogRecord) error {
	return l.repo.Insert(logs)
}

func (l *LogUseCase) Find(dateFrom time.Time, dateTo time.Time, limit uint) (records []entity.LogRecord, limited bool, err error) {
	r, lim, e := l.repo.Find(dateFrom, dateTo, limit)
	return r, lim, e
}
