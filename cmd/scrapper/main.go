package main

import (
	"context"
	"fmt"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	"github.com/go-co-op/gocron"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
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
	minCons          = 4
	maxCons          = 15
)

func main() {
	var logLevel = new(slog.LevelVar)

	logLevel.Set(slog.LevelDebug)

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))
	dbConfig, err := sql.NewConfig()

	if err != nil {
		logger.Debug("ошибка при получении конфига БД", "err", err.Error())
	}

	pgxPool, err := initPgxPool(dbConfig)

	if err != nil {
		logger.Debug("ошибка инициализации пула соединений", "err", err.Error())
	}

	var userStore scrapservice.UserRepo

	switch dbConfig.AccessType {
	case "SQL":
		cleanSqlStore := cleansql.NewStore(dbConfig, pgxPool)

		userStore = cleanSqlStore

	case "ORM":
		builderSqlStore := buildersql.NewStore(dbConfig, pgxPool)

		userStore = builderSqlStore
	default:
		logger.Debug("ошибка конфигурации", "err", "переменная окружения AccessType должна быть SQL или ORM")
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

func initPgxPool(dbConfig *sql.DBConfig) (*pgxpool.Pool, error) {
	pgxConfig, err := pgxpool.ParseConfig(dbConfig.ToDSN())

	if err != nil {
		return nil, fmt.Errorf("ошибка при парсинга строки подключения к БД: %w", err)
	}

	pgxConfig.MinConns, pgxConfig.MaxConns = minCons, maxCons // разобраться с тем как настроить время ожидания конекшена

	pgxPool, err := pgxpool.NewWithConfig(context.Background(), pgxConfig)

	if err != nil {
		return nil, fmt.Errorf("ошибка при создании pgxpool.Pool: %w", err)
	}

	if err = pgxPool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("ошибка при проверке соединения с БД: %w", err)
	}

	return pgxPool, nil
}
