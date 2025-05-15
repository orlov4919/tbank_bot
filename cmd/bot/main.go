package main

import (
	"context"
	"github.com/go-co-op/gocron"
	"linkTraccer/internal/application/botservice"
	"linkTraccer/internal/infrastructure/botconf"
	"linkTraccer/internal/infrastructure/bothandler"
	"linkTraccer/internal/infrastructure/cache/redisstore"
	"linkTraccer/internal/infrastructure/database/file/contextstorage"
	"linkTraccer/internal/infrastructure/kafka/consumer"
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

	appConf, err := botconf.New()
	if err != nil {
		logger.Error("ошибка при получении конфига бота", "err", err.Error())
		return
	}

	tgClient := telegram.NewClient(&http.Client{Timeout: time.Minute}, appConf.BotToken, telegramBotAPI)
	ctxStore := contextstorage.New()
	scrapClient := scrapclient.New(&http.Client{Timeout: time.Minute}, appConf.ScrapperHost, appConf.ScrapperPort)

	redisConf, err := redisstore.NewConfig()
	if err != nil {
		logger.Error("ошибка при создании конфига редис", "err", err.Error())
		return
	}

	redisStore, err := redisstore.NewStore(redisConf)
	if err != nil {
		logger.Error("ошибка при создании redis хранилища", "err", err.Error())
		return
	}

	tgBot := botservice.New(tgClient, scrapClient, ctxStore, redisStore, logger, appConf.BotBatch)

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

	go startReceiveUpdates(context.Background(), tgClient, appConf, logger, wg)

	wg.Wait()
}

func startReceiveUpdates(ctx context.Context, tg botservice.TgClient, config *botconf.Config, logger *slog.Logger, wg *sync.WaitGroup) {
	defer wg.Done()

	switch config.UpdatesTransport {
	case "KAFKA":
		logger.Info("запущен консьюмер принимающий обновления по ссылкам")
		initAndRunConsumer(ctx, tg, config, logger)
	case "HTTP":
		logger.Info("запущен сервер принимающий обновления по ссылкам")
		initAndRunServer(tg, config, logger)
	default:
		logger.Error("ошибка конфигурации", "err", "получение обновлений должно быть KAFKA или HTTP")
	}
}

func initAndRunServer(tg botservice.TgClient, config *botconf.Config, logger *slog.Logger) {
	r := mux.NewRouter()

	r.HandleFunc("/updates", bothandler.New(tg, logger).HandleLinkUpdates).Methods(http.MethodPost)

	srv := &http.Server{
		Addr:         config.BotPort,
		Handler:      r,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	if err := srv.ListenAndServe(); err != nil {
		logger.Error("сервер принимающий обновления ссылок, закончил работу", "err", err.Error())
	}
}

func initAndRunConsumer(ctx context.Context, tg botservice.TgClient, config *botconf.Config, logger *slog.Logger) {
	conf, err := consumer.NewConfig()
	if err != nil {
		logger.Error("ошибка при создании конфига kafka консьюмера", "err", err.Error())
	}

	consumer := consumer.New(tg, conf, logger)

	err = consumer.ReadUserUpdates(ctx)
	if err != nil {
		logger.Error("консьюмер закончил работу с ошибкой", "err", err.Error())
	}

	err = consumer.Close()
	if err != nil {
		logger.Error("ошибка при закрытии консьюмера", "err", err.Error())
	}
}
