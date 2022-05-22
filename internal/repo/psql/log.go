// Package psql Содержит реализацию интерфейса репозитория логов для postgresql
package psql

import (
	"context"
	"fmt"
	"time"

	"github.com/n-r-w/log-server-v2/internal/domain/entity"
	"github.com/n-r-w/log-server-v2/pkg/postgres"
)

type logRepo struct {
	*postgres.Postgres
	maxLogRecordsResult int
}

func NewLog(pg *postgres.Postgres, maxLogRecordsResult int) *logRepo {
	return &logRepo{
		Postgres:            pg,
		maxLogRecordsResult: maxLogRecordsResult,
	}
}

func (p *logRepo) Insert(records []entity.LogRecord) error {
	var sqlText string

	for _, lr := range records {
		if err := lr.Validate(); err != nil {
			return err
		}

		t, _ := lr.LogTime.UTC().MarshalText()
		sqlText += fmt.Sprintf(`INSERT INTO log (record_timestamp, level, message1, message2, message3) 
		 					    VALUES ('%s', %d, '%s', '%s', '%s');`,
			t, lr.Level, lr.Message1, lr.Message2, lr.Message3)
	}

	_, err := p.Pool.Exec(context.Background(), sqlText)

	return err
}

func (p *logRepo) Find(dateFrom time.Time, dateTo time.Time, limit int) (records []entity.LogRecord, limited bool, err error) {
	rows, err := p.Pool.Query(context.Background(),
		`SELECT id, record_timestamp, real_timestamp, level,  message1, COALESCE(message2, ''), COALESCE(message3, '') 
		FROM log
		WHERE ($1 OR record_timestamp >= $2) AND ($3 OR record_timestamp <= $4)
		ORDER BY record_timestamp DESC
		LIMIT $5`,
		dateFrom.IsZero(), dateFrom, dateTo.IsZero(), dateTo, limit+1)
	if err != nil {
		return nil, false, err
	}
	defer rows.Close() // освобождаем контекст sql запроса при выходе

	var recs []entity.LogRecord

	var rowCount uint64
	limited = false

	for rows.Next() {
		var record entity.LogRecord

		if err := rows.Scan(&record.ID, &record.LogTime, &record.RealTime,
			&record.Level, &record.Message1, &record.Message2, &record.Message3); err != nil {
			return nil, false, err
		}

		rowCount++
		if rowCount > uint64(limit) {
			limited = true

			break
		}

		if rowCount > uint64(p.maxLogRecordsResult) {
			err := fmt.Errorf("too many records, max %d", p.maxLogRecordsResult)

			return nil, false, err
		}

		recs = append(recs, record)
	}

	// при rows.Scan может быть ошибка и тогда defer rows.Close() не вызовется
	// поэтому надежнее сделать как defer rows.Close(), так и прямое закрытие здесь
	rows.Close()

	return recs, limited, nil
}
