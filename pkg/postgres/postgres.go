package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/n-r-w/log-server-v2/pkg/logger"
)

const (
	defaultMaxConns           = 10
	defaultmaxPoolSize        = 10
	defaultConnAttempts       = 10
	defaultMaxMaxConnIdleTime = time.Second
	defaultConnTimeout        = time.Second * 5
	defaultStatementTimeout   = time.Second * 5
	defaultReconnectTimeout   = time.Second
)

type Postgres struct {
	maxConns           int
	maxPoolSize        int
	connAttempts       int
	maxMaxConnIdleTime time.Duration
	connTimeout        time.Duration
	statementTimeout   time.Duration
	reconnectTimeout   time.Duration

	Pool *pgxpool.Pool
}

func New(url string, logger logger.Interface, options ...Option) (*Postgres, error) {
	pg := &Postgres{
		maxConns:           defaultMaxConns,
		maxPoolSize:        defaultmaxPoolSize,
		connAttempts:       defaultConnAttempts,
		maxMaxConnIdleTime: defaultMaxMaxConnIdleTime,
		connTimeout:        defaultConnTimeout,
		statementTimeout:   defaultStatementTimeout,
		reconnectTimeout:   defaultReconnectTimeout,
	}

	for _, opt := range options {
		opt(pg)
	}

	poolConfig, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, fmt.Errorf("pgxpool parse config error: %v", err)
	}

	poolConfig.MaxConns = int32(pg.maxPoolSize)
	poolConfig.MaxConnIdleTime = pg.maxMaxConnIdleTime
	poolConfig.MaxConns = int32(pg.maxConns)

	connAttempts := pg.connAttempts
	for connAttempts > 0 {
		pg.Pool, err = pgxpool.ConnectConfig(context.Background(), poolConfig)
		if err == nil {
			break
		}

		logger.Info("Postgres is trying to connect, attempts left: %d", connAttempts)

		time.Sleep(pg.reconnectTimeout)

		connAttempts--
	}

	if err != nil {
		return nil, fmt.Errorf("postgres connection error: %v", err)
	}

	return pg, nil
}

func (p *Postgres) Close() {
	if p.Pool != nil {
		p.Pool.Close()
	}
}
