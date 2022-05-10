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
	"github.com/n-r-w/log-server-v2/pkg/httpserver"
	"github.com/n-r-w/log-server-v2/pkg/logger"
	"github.com/n-r-w/log-server-v2/pkg/postgres"
)

func Start(cfg *config.Config, logger logger.Interface) {
	// создаем доступ к БД
	pg, err := postgres.New(cfg.DatabaseURL, logger)
	if err != nil {
		logger.Error("postgress error: %v", err)

		return
	}

	// создаем репозитории
	userRepo := psql.NewUser(pg, logger, cfg.SuperAdminID, cfg.SuperAdminLogin, cfg.SuperPassword,
		cfg.PasswordRegex, cfg.PasswordRegexError)
	logRepo := psql.NewLog(pg, cfg.MaxLogRecordsResult)

	// создаем сценарии
	userCase := usecase.NewUserCase(userRepo, cfg.SuperAdminID)
	logCase := usecase.NewLogCase(logRepo)

	// создаем маршрутизатор запросов
	rt := router.NewRouter(logger, userCase, logCase, cfg.SessionEncriptionKey, cfg.SuperAdminID, cfg.SessionAge, cfg.MaxLogRecordsResult)

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
	if err != nil {
		logger.Error("shutdown error: %v", err)
	} else {
		logger.Info("shutdown ok")
	}
}
