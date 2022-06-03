package main

import (
	"flag"

	"github.com/n-r-w/log-server-v2/internal/app"
	"github.com/n-r-w/log-server-v2/internal/config"
	"github.com/n-r-w/log-server-v2/pkg/logger"
)

func main() {
	lg := logger.New()

	var configPath string
	// описание флагов командной строки
	flag.StringVar(&configPath, "config-path", "", "path to config file")

	// обработка командной строки
	flag.Parse()

	// читаем конфиг
	cfg, err := config.New(configPath, lg)
	if err != nil {
		lg.Fatal("read config error: %v", err)

		return
	}

	app.Start(cfg, lg)
}
