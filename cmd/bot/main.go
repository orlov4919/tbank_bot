package main

import (
	"linkTraccer/internal/application/botservice"
	"linkTraccer/internal/infrastructure/botconfig"
	"linkTraccer/internal/infrastructure/bothandler"
	"linkTraccer/internal/infrastructure/database/file/contextstorage"
	"linkTraccer/internal/infrastructure/scrapperclient"
	"linkTraccer/internal/infrastructure/telegram"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/go-co-op/gocron"
)

const (
	telegramBotAPI = "api.telegram.org"
)

func main() {
	var logLevel = new(slog.LevelVar)

	logLevel.Set(slog.LevelInfo)

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))
	config, err := botconfig.New()

	if err != nil {
		logger.Debug("ошибка при настройке конфига", "err", err.Error())
	}

	tgClient := telegram.NewClient(&http.Client{Timeout: time.Minute}, config.Token, telegramBotAPI)
	ctxStore := contextstorage.New()
	scrapClient := scrapperclient.New(&http.Client{Timeout: time.Minute}, config.ScrapperHost, config.ScrapperPort)
	tgBot := botservice.New(tgClient, scrapClient, ctxStore, logger, 5)

	tgBot.Init()

	s := gocron.NewScheduler(time.UTC)

	_, err = s.Every(time.Second * 3).Do(tgBot.CheckUsersMsg)

	if err != nil {
		logger.Debug("ошибка в работе планировщика", "err", err.Error())
	}

	s.StartAsync()

	mux := http.NewServeMux()

	mux.HandleFunc("/updates", bothandler.New(tgClient, logger).HandleLinkUpdates)

	srv := &http.Server{
		Addr:         config.BotPort,
		Handler:      mux,
		ReadTimeout:  30 * time.Second, // Таймаут чтения запроса
		WriteTimeout: 30 * time.Second, // Таймаут записи ответа
		IdleTimeout:  30 * time.Second, // Таймаут простоя соединения
	}

	_ = srv.ListenAndServe()
}
