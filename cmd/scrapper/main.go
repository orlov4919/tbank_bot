package main

import (
	"linkTraccer/internal/application/scrapperservice"
	"linkTraccer/internal/infrastructure/botclient"
	"linkTraccer/internal/infrastructure/database/file/userstorage"
	"linkTraccer/internal/infrastructure/scrapperconfig"
	"linkTraccer/internal/infrastructure/scrapperhandlers"
	"linkTraccer/internal/infrastructure/siteclients/github"
	"linkTraccer/internal/infrastructure/siteclients/stackoverflow"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/go-co-op/gocron"
)

const (
	stackOverflowAPI = "api.stackexchange.com"
	gitHubAPI        = "api.github.com"
)

func main() {
	var logLevel = new(slog.LevelVar)

	logLevel.Set(slog.LevelInfo)

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))

	config, err := scrapperconfig.New()

	if err != nil {
		logger.Debug("ошибка при настройке конфига", "err", err.Error())
	}

	stackClient := stackoverflow.NewClient(stackOverflowAPI, &http.Client{Timeout: time.Second * 10})
	tgBotClient := botclient.New(config.TgBotServerURL, &http.Client{Timeout: time.Second * 10})
	userRepo := userstorage.NewFileStorage()
	gitClient := github.NewClient(gitHubAPI, config.GitHubToken, &http.Client{Timeout: time.Minute})

	scrapper := scrapperservice.New(userRepo, tgBotClient, logger, stackClient, gitClient)

	s := gocron.NewScheduler(time.UTC)

	_, err = s.Every(time.Minute).Do(scrapper.CheckLinkUpdates)

	if err != nil {
		logger.Debug("ошибка при запуске планировщика", "err", err.Error())
	}

	s.StartAsync()

	mux := http.NewServeMux()
	linksHandler := scrapperhandlers.NewLinkHandler(userRepo, logger, stackClient, gitClient)
	chatHandler := scrapperhandlers.NewChatHandler(userRepo, logger)

	mux.HandleFunc("/tg-chat/", chatHandler.HandleChatChanges)
	mux.HandleFunc("/links", linksHandler.HandleLinksChanges)

	srv := &http.Server{
		Addr:         config.ScrapperServerPort,
		Handler:      mux,
		ReadTimeout:  30 * time.Second, // Таймаут чтения запроса
		WriteTimeout: 30 * time.Second, // Таймаут записи ответа
		IdleTimeout:  30 * time.Second, // Таймаут простоя соединения
	}

	_ = srv.ListenAndServe()
}
