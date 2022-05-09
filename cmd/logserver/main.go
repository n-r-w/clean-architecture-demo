package main

import (
	"flag"

	"github.com/n-r-w/log-server-v2/internal/app"
	"github.com/n-r-w/log-server-v2/internal/config"
	"github.com/n-r-w/log-server-v2/pkg/logger"
)

func main() {
	logger := logger.New()

	var configPath string
	// описание флагов командной строки
	flag.StringVar(&configPath, "config-path", "config/server.toml", "path to config file")

	// обработка командной строки
	flag.Parse()

	// читаем конфиг
	config, err := config.New(configPath)
	if err != nil {
		logger.Fatal("read config error %v", err)

		return
	}

	app.Start(config, logger)
}
