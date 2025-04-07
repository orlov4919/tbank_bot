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
	"linkTraccer/internal/infrastructure/database/sql/transactor"
	"linkTraccer/internal/infrastructure/scrapconfig"
	"linkTraccer/internal/infrastructure/scraphandlers"
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
	logLevel := new(slog.LevelVar)

	logLevel.Set(slog.LevelInfo)

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))

	dbConfig, err := sql.NewConfig()

	if err != nil {
		logger.Error("ошибка при получении конфига БД", "err", err.Error())

		return
	}

	pgxPool, err := initPgxPool(dbConfig)

	if err != nil {
		logger.Error("ошибка инициализации пула соединений", "err", err.Error())

		return
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
		logger.Error("ошибка конфигурации", "err", "переменная окружения AccessType должна быть SQL или ORM")

		return
	}

	config, err := scrapconfig.New()

	if err != nil {
		logger.Error("ошибка при получении конфига scrapper", "err", err.Error())

		return
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
		logger.Error("ошибка при запуске планировщика с проверкой ссылок", "err", err.Error())
	}

	s.StartAsync()

	r := mux.NewRouter()

	transacter := transactor.New(pgxPool)

	linksHandler := scraphandlers.NewLinkHandler(userStore, transacter, logger, stackClient, gitClient)
	chatHandler := scraphandlers.NewChatHandler(userStore, transacter, logger)

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

	pgxConfig.MinConns, pgxConfig.MaxConns = minCons, maxCons // разобраться с тем как настроить время ожидания конекта

	pgxPool, err := pgxpool.NewWithConfig(context.Background(), pgxConfig)

	if err != nil {
		return nil, fmt.Errorf("ошибка при создании pgxpool.Pool: %w", err)
	}

	if err = pgxPool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("ошибка при проверке соединения с БД: %w", err)
	}

	return pgxPool, nil
}

//func initUserStore(dbConfig *sql.DBConfig, pgxPool *pgxpool.Pool) (scrapservice.UserRepo, error) {
//	var userStore scrapservice.UserRepo
//
//	switch dbConfig.AccessType {
//	case "SQL":
//		cleanSqlStore := cleansql.NewStore(dbConfig, pgxPool)
//
//		userStore = cleanSqlStore
//
//	case "ORM":
//		builderSqlStore := buildersql.NewStore(dbConfig, pgxPool)
//
//		userStore = builderSqlStore
//	default:
//		logger.Error("ошибка конфигурации", "err", "переменная окружения AccessType должна быть SQL или ORM")
//
//		return
//	}
//
//}
