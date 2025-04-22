package main

import (
	"github.com/go-co-op/gocron"
	"linkTraccer/internal/application/botservice"
	"linkTraccer/internal/infrastructure/botconfig"
	"linkTraccer/internal/infrastructure/bothandler"
	"linkTraccer/internal/infrastructure/database/file/contextstorage"
	"linkTraccer/internal/infrastructure/scrapclient"
	"linkTraccer/internal/infrastructure/telegram"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gorilla/mux"
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
		logger.Error("ошибка при получении конфига бота", "err", err.Error())
		return
	}

	tgClient := telegram.NewClient(&http.Client{Timeout: time.Minute}, config.Token, telegramBotAPI)
	ctxStore := contextstorage.New()
	scrapClient := scrapclient.New(&http.Client{Timeout: time.Minute}, config.ScrapperHost, config.ScrapperPort)
	tgBot := botservice.New(tgClient, scrapClient, ctxStore, logger, 5)

	if err = tgBot.Init(); err != nil {
		logger.Error("ошибка при инициализации бота", "err", err.Error())
		return
	}

	logger.Info("инициализация телеграмм бота прошла успешно")

	s := gocron.NewScheduler(time.UTC)

	_, err = s.Every(time.Second * 3).Do(tgBot.ProcessMsg)
	if err != nil {
		logger.Error("ошибка в работе планировщика", "err", err.Error())
		return
	}

	s.StartAsync()
	logger.Info("планировщик с проверкой новых сообщений в боте, успешно запущен")

	wg := &sync.WaitGroup{}

	wg.Add(1)

	go func() {
		defer wg.Done()
		initAndRunServer(tgClient, config, logger)
	}()

	logger.Info("запущен сервер принимающий обновления по ссылкам")
	wg.Wait()
}

func initAndRunServer(tgClient botservice.TgClient, config *botconfig.Config, logger *slog.Logger) {
	r := mux.NewRouter()

	r.HandleFunc("/updates", bothandler.New(tgClient, logger).HandleLinkUpdates).Methods(http.MethodPost)

	srv := &http.Server{
		Addr:         config.BotPort,
		Handler:      r,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	if err := srv.ListenAndServe(); err != nil {
		logger.Error("сервер закончил работу", "err", err.Error())
	}
}

func initAndRunConsumer() {

}
