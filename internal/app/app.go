// Package app ...
package app

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/n-r-w/log-server-v2/internal/config"
	"github.com/n-r-w/log-server-v2/internal/domain/usecase"
	"github.com/n-r-w/log-server-v2/internal/presentation/http/router"
	"github.com/n-r-w/log-server-v2/internal/repo/psql"
	"github.com/n-r-w/log-server-v2/internal/repo/wbuf"
	"github.com/n-r-w/log-server-v2/pkg/httpserver"
	"github.com/n-r-w/log-server-v2/pkg/logger"
	"github.com/n-r-w/log-server-v2/pkg/postgres"
)

func Start(cfg *config.Config, logger logger.Interface) {
	// создаем доступ к БД
	pg, err := postgres.New(cfg.DatabaseURL, logger,
		postgres.MaxConns(cfg.MaxDbSessions),
		postgres.MaxMaxConnIdleTime(time.Duration(cfg.MaxDbSessionIdleTimeSec)*time.Second))
	if err != nil {
		logger.Error("postgress error: %v", err)

		return
	}

	// создаем репозитории
	userRepo := psql.NewUser(pg, logger, uint64(cfg.SuperAdminID), cfg.SuperAdminLogin, cfg.SuperPassword,
		cfg.PasswordRegex, cfg.PasswordRegexError)
	logRepo := psql.NewLog(pg, cfg.MaxLogRecordsResult)

	// создаем буфер для асинхронной записи в БД
	buffer := wbuf.NewDispatcher(cfg.MaxDbSessions, cfg.RateLimit, cfg.RateLimitBurst, logRepo, logger)

	// создаем сценарии
	userCase := usecase.NewUserCase(userRepo, uint64(cfg.SuperAdminID))
	logCase := usecase.NewLogCase(buffer) // вместо logRepo передаем буфер, т.к. он реализует интерфейс usecase.LogInterface

	// создаем маршрутизатор запросов
	rt := router.NewRouter(logger, userCase, logCase, cfg.SessionEncriptionKey, uint64(cfg.SuperAdminID), cfg.SessionAge, cfg.MaxLogRecordsResult)

	// запускаем http сервер
	httpServer := httpserver.New(rt.Handler(), logger,
		httpserver.Address(cfg.Host, cfg.Port),
		httpserver.ReadTimeout(time.Second*time.Duration(cfg.HttpReadTimeout)),
		httpserver.WriteTimeout(time.Second*time.Duration(cfg.HttpWriteTimeout)),
		httpserver.ShutdownTimeout(time.Second*time.Duration(cfg.HttpShutdownTimeout)),
	)

	// и ждем от него сигнала
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	select {
	case <-interrupt:
		logger.Info("shutdown, timeout %d ...", cfg.HttpShutdownTimeout)
	case err = <-httpServer.Notify():
		logger.Error("http server notification: %v", err)
	}

	// ждем завершения
	err = httpServer.Shutdown()
	buffer.Stop()
	if err != nil {
		logger.Error("shutdown error: %v", err)
	} else {
		logger.Info("shutdown ok")
	}

}
