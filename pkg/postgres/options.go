package postgres

import "time"

type Option func(*Postgres)

func MaxConns(n int) Option {
	return func(p *Postgres) {
		p.maxConns = n
	}
}

func ConnAttempts(n int) Option {
	return func(p *Postgres) {
		p.connAttempts = n
	}
}

func MaxMaxConnIdleTime(n time.Duration) Option {
	return func(p *Postgres) {
		p.maxMaxConnIdleTime = n
	}
}

func ConnTimeout(n time.Duration) Option {
	return func(p *Postgres) {
		p.connTimeout = n
	}
}

func StatementTimeout(n time.Duration) Option {
	return func(p *Postgres) {
		p.statementTimeout = n
	}
}

func ReconnectTimeout(n time.Duration) Option {
	return func(p *Postgres) {
		p.reconnectTimeout = n
	}
}
