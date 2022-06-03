// Package wbuf - асинхронная запись в БД, чтобы не тормозить обработку веб запросов
package wbuf

import (
	"fmt"
	"time"

	"github.com/gammazero/workerpool"
	"github.com/n-r-w/log-server-v2/internal/domain/entity"
	"github.com/n-r-w/log-server-v2/internal/domain/usecase"
	"github.com/n-r-w/log-server-v2/pkg/logger"
	"golang.org/x/time/rate"
)

type Dispatcher struct {
	log     logger.Interface
	dbRepo  usecase.LogInterface
	limiter *rate.Limiter
	pool    *workerpool.WorkerPool
}

func NewDispatcher(workerCount int, rateLimit int, rateLimitBurst int, dbRepo usecase.LogInterface, log logger.Interface) *Dispatcher {
	d := &Dispatcher{
		log:     log,
		dbRepo:  dbRepo,
		limiter: rate.NewLimiter(rate.Limit(rateLimit), rateLimitBurst),
		pool:    workerpool.New(workerCount),
	}

	// Вывод в фоновом режиме информации о размере буфера раз в секунду
	go func() {
		for {
			size := d.pool.WaitingQueueSize()
			if size > d.pool.Size()*2 {
				d.log.Info("queue size: %d, pool size: %d", size, d.dbRepo.PoolSize())
			}
			time.Sleep(time.Second)
		}
	}()

	return d
}

func (d *Dispatcher) PoolSize() int {
	return d.dbRepo.PoolSize()
}

// Insert - реализация интерфейса usecase.LogInterface
func (d *Dispatcher) Insert(records []entity.LogRecord) error {
	// Защита от DDOS и в целом от перегрузки сервера БД запросами
	if !d.limiter.Allow() {
		return fmt.Errorf("too many requests")
	}
	// Контроль за размером очереди пула задач. Если дать ему бескотрольно расти, то можно остаться без свободных ресурсов
	// Фактически тут мы искусственно увеличиваем время отклика входящих запросов при переполнении очереди задач
	for {
		if d.pool.WaitingQueueSize() > d.pool.Size()*2 {
			time.Sleep(time.Millisecond)
		} else {
			break
		}
	}
	// Отправляем задачу на асинхронную обработку
	d.pool.Submit(func() {
		if err := d.dbRepo.Insert(records); err != nil {
			d.log.Error("worker error: %v", err)
		}
	})

	return nil
}

// Find - реализация интерфейса usecase.LogInterface для его подмены
func (d *Dispatcher) Find(dateFrom time.Time, dateTo time.Time, limit int) (records []entity.LogRecord, limited bool, err error) {
	// просто пересылаем запрос
	return d.dbRepo.Find(dateFrom, dateTo, limit)
}

func (d *Dispatcher) Stop() {
	d.log.Info("buffer dispatcher stoping...")
	d.pool.StopWait()
	d.log.Info("buffer dispatcher stopped OK")
}
