package main

import (
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	"github.com/go-co-op/gocron"
	"github.com/gorilla/mux"
	"linkTraccer/internal/application/scrapper/notifiers/tgnotifier"
	"linkTraccer/internal/application/scrapper/scrapservice"
	"linkTraccer/internal/infrastructure/botclient"
	"linkTraccer/internal/infrastructure/database/sql"
	"linkTraccer/internal/infrastructure/database/sql/buildersql"
	"linkTraccer/internal/infrastructure/database/sql/cleansql"
	"linkTraccer/internal/infrastructure/scrapperconfig"
	"linkTraccer/internal/infrastructure/scrapperhandlers"
	"linkTraccer/internal/infrastructure/siteclients/github"
	"linkTraccer/internal/infrastructure/siteclients/stackoverflow"
	"log/slog"
	"net/http"
	"os"
	"time"
)

const (
	stackOverflowAPI = "api.stackexchange.com"
	gitHubAPI        = "api.github.com"
	maxPreviewLen    = 200
)

func main() {
	var logLevel = new(slog.LevelVar)

	logLevel.Set(slog.LevelDebug)

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))

	dbConfig, err := sql.NewConfig()

	if err != nil {
		logger.Debug("ошибка при получении конфига БД", "err", err.Error())
	}

	var userStore scrapservice.UserRepo

	initStore := func(store interface{ Open() error }) error {
		if err := store.Open(); err != nil {
			return err
		}

		return nil
	}

	switch dbConfig.AccessType {
	case "SQL":
		cleanSqlStore := cleansql.NewStore(dbConfig)
		if err = initStore(cleanSqlStore); err != nil {
			logger.Debug("ошибка при запуске БД", "err", err.Error())
		}

		userStore = cleanSqlStore

	case "ORM":
		builderSqlStore := buildersql.NewStore(dbConfig)

		if err = initStore(builderSqlStore); err != nil {
			logger.Debug("ошибка при запуске БД", "err", err.Error())
		}

		userStore = builderSqlStore
	default:
		logger.Debug("ошибка при запуске БД", "err", "AccessType должен быть SQL или ORM")
	}

	config, err := scrapperconfig.New()

	if err != nil {
		logger.Debug("ошибка при получении конфига scrapper", "err", err.Error())
	}

	stackClient := stackoverflow.NewClient(stackOverflowAPI, &http.Client{Timeout: time.Second * 10},
		stackoverflow.HTMLStrCleaner(maxPreviewLen))
	gitClient := github.NewClient(gitHubAPI, config.GitHubAPIKey, &http.Client{Timeout: time.Second * 10})
	tgBotClient := botclient.New(config.BotHost+config.BotPort, &http.Client{Timeout: time.Second * 10})
	notifierService := tgnotifier.New(userStore, tgBotClient)
	scrapper := scrapservice.New(userStore, notifierService, logger, stackClient, gitClient)

	s := gocron.NewScheduler(time.UTC)

	_, err = s.Every(time.Minute).Do(scrapper.CheckLinksUpdates)

	if err != nil {
		logger.Debug("ошибка при запуске планировщика с проверкой ссылок", "err", err.Error())
	}

	s.StartAsync()

	r := mux.NewRouter()

	linksHandler := scrapperhandlers.NewLinkHandler(userStore, logger, stackClient, gitClient)
	chatHandler := scrapperhandlers.NewChatHandler(userStore, logger)

	r.HandleFunc("/tg-chat/{id}", chatHandler.HandleChatChanges).
		Methods(http.MethodPost, http.MethodDelete)
	r.HandleFunc("/links", linksHandler.HandleLinksChanges).
		Methods(http.MethodGet, http.MethodPost, http.MethodDelete)

	srv := &http.Server{
		Addr:         config.ScrapperPort,
		Handler:      r,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	_ = srv.ListenAndServe()
}
