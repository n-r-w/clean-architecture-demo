// Package wbuf - асинхронная запись в БД, чтобы не тормозить обработку веб запросов
package wbuf

import (
	"time"

	"github.com/gammazero/workerpool"
	"github.com/n-r-w/log-server-v2/internal/domain/entity"
	"github.com/n-r-w/log-server-v2/internal/domain/usecase"
	"github.com/n-r-w/log-server-v2/pkg/logger"
)

type Dispatcher struct {
	log      logger.Interface
	dbRepo   usecase.LogInterface
	stepSize uint8
	pool     *workerpool.WorkerPool
}

func NewDispatcher(workerCount uint16, dbRepo usecase.LogInterface, log logger.Interface) *Dispatcher {
	d := &Dispatcher{
		log:    log,
		dbRepo: dbRepo,
		pool:   workerpool.New(int(workerCount)),
	}

	return d
}

// Insert - реализация интерфейса usecase.LogInterface
func (d *Dispatcher) Insert(records []entity.LogRecord) error {
	d.pool.Submit(func() {
		if err := d.dbRepo.Insert(records); err != nil {
			d.log.Error("worker error: %v", err)
		}
	})

	return nil
}

// Find - реализация интерфейса usecase.LogInterface для его подмены
func (d *Dispatcher) Find(dateFrom time.Time, dateTo time.Time, limit uint) (records []entity.LogRecord, limited bool, err error) {
	// просто пересылаем запрос
	return d.dbRepo.Find(dateFrom, dateTo, limit)
}

func (d *Dispatcher) Stop() {
	d.log.Info("buffer dispatcher stoping...")
	d.pool.StopWait()
	d.log.Info("buffer dispatcher stopped OK")
}
