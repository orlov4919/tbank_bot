package main

import (
	"context"
	"errors"
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
	"linkTraccer/internal/infrastructure/kafka/producer"
	"linkTraccer/internal/infrastructure/scrapconfig"
	"linkTraccer/internal/infrastructure/scraphandlers"
	"linkTraccer/internal/infrastructure/siteclients/github"
	"linkTraccer/internal/infrastructure/siteclients/stackoverflow"
	"log/slog"
	"net"
	"net/http"
	"os"
	"sync"
	"time"
)

const (
	stackOverflowAPI = "api.stackexchange.com"
	gitHubAPI        = "api.github.com"
	maxPreviewLen    = 200
	minCons          = 4
	maxCons          = 15
)

type UserRepo = scrapservice.UserRepo
type Transactor = scrapservice.Transactor
type SiteClient = scrapservice.SiteClient
type Config = scrapconfig.Config

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

	logger.Info("соединение с БД успешно установлено")

	userStore, err := initStore(dbConfig, pgxPool)
	if err != nil {
		logger.Error("ошибка при создании хранилища", "err", err.Error())
		return
	}

	config, err := scrapconfig.New()
	if err != nil {
		logger.Error("ошибка при получении конфига scrapper", "err", err.Error())
		return
	}

	dbTransactor := transactor.New(pgxPool)
	stackClient := stackoverflow.NewClient(stackOverflowAPI, &http.Client{Timeout: time.Second * 10},
		stackoverflow.HTMLStrCleaner(maxPreviewLen))
	gitClient := github.NewClient(gitHubAPI, config.GitHubAPIKey, &http.Client{Timeout: time.Second * 10})

	tgBotClient, err := initUpdatesTransport(config)
	if err != nil {
		logger.Error("ошибка при инициализации клиента тг бота", "err", err.Error())
		return
	}

	notifierService := tgnotifier.New(userStore, tgBotClient)
	scrapper := scrapservice.New(userStore, notifierService, logger, stackClient, gitClient)
	scheduler := gocron.NewScheduler(time.UTC)

	_, err = scheduler.Every(time.Minute).Do(scrapper.LinksUpdates)
	if err != nil {
		logger.Error("ошибка при запуске планировщика с проверкой ссылок", "err", err.Error())
		return
	}

	scheduler.StartAsync()

	logger.Info("планировщик с проверкой ссылок успешно запущен")
	logger.Info("сервер успешно запущен")

	wg := &sync.WaitGroup{}

	wg.Add(1)

	go func() {
		defer wg.Done()
		initAndRunServer(userStore, dbTransactor, logger, config, gitClient, stackClient)
	}()

	wg.Wait()
}

func initUpdatesTransport(config *scrapconfig.Config) (tgnotifier.BotClient, error) {
	switch config.UpdatesTransport {
	case "KAFKA":
		kafkaConfig, err := producer.NewConfig()
		if err != nil {
			return nil, err
		}

		return producer.New(kafkaConfig), nil

	case "HTTP":
		httpConfig, err := botclient.NewConfig()
		if err != nil {
			return nil, err
		}

		return botclient.New(net.JoinHostPort(httpConfig.BotHost, httpConfig.BotPort), &http.Client{Timeout: time.Second * 10}), nil
	default:
		return nil, errors.New("UPDATE_TRANSPORT должен быть KAFKA или HTTP")
	}
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

func initStore(config *sql.DBConfig, pool *pgxpool.Pool) (UserRepo, error) {
	switch config.AccessType {
	case "SQL":
		return cleansql.NewStore(config, pool), nil
	case "ORM":
		return buildersql.NewStore(config, pool), nil
	default:
		return nil, errors.New("переменная окружения AccessType, должна быть SQL или ORM")
	}
}

func initAndRunServer(userStore UserRepo, dbTransactor Transactor, log *slog.Logger, cfg *Config, siteClients ...SiteClient) {
	r := mux.NewRouter()
	linksHandler := scraphandlers.NewLinkHandler(userStore, dbTransactor, log, siteClients...)
	chatHandler := scraphandlers.NewChatHandler(userStore, dbTransactor, log)

	r.HandleFunc("/tg-chat/{id}", chatHandler.HandleChatChanges).
		Methods(http.MethodPost, http.MethodDelete)
	r.HandleFunc("/links", linksHandler.HandleLinksChanges).
		Methods(http.MethodGet, http.MethodPost, http.MethodDelete)

	srv := &http.Server{
		Addr:         cfg.ScrapperPort,
		Handler:      r,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	err := srv.ListenAndServe()
	if err != nil {
		log.Error("сервер закончил работу", "err", err.Error())
	}
}
