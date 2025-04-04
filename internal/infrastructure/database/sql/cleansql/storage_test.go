package cleansql_test

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"linkTraccer/internal/domain/scrapper"
	"linkTraccer/internal/infrastructure/database/sql"
	"linkTraccer/internal/infrastructure/database/sql/cleansql"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	dbName         = "linkTraccer"
	dbUser         = "postgres"
	dbPassword     = "1234"
	migrationsPath = "../../../../../migrations"
	bathSize       = 2

	githubLink        = "https://github.com/orlov4919/test"
	stackoverflowLink = "https://stackoverflow.com/"
	firstID           = 1
	secondID          = 2
	thirdID           = 3
)

type Link = scrapper.Link

func SetupContainer() (*postgres.PostgresContainer, error) {
	postgresContainer, err := postgres.Run(context.Background(),
		"postgres:16-alpine",
		postgres.WithDatabase(dbName),
		postgres.WithUsername(dbUser),
		postgres.WithPassword(dbPassword),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(10*time.Second)),
	)

	return postgresContainer, err
}

func DatabaseURL(container *postgres.PostgresContainer) (string, error) {

	connString, err := container.ConnectionString(context.Background())

	if err != nil {
		return connString, err
	}

	return connString + "sslmode=disable", nil

}

func RunMigrations(dbURL, migrationsPath string) error {
	m, err := migrate.New("file://"+migrationsPath, dbURL)

	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	srcErr, dbErr := m.Close()

	return errors.Join(srcErr, dbErr)
}

func ConnectToDB(dbURL string) (*pgxpool.Pool, error) {
	pgxConfig, err := pgxpool.ParseConfig(dbURL)

	if err != nil {
		return nil, err
	}

	pgxPool, err := pgxpool.NewWithConfig(context.Background(), pgxConfig)

	if err != nil {
		return nil, err
	}

	if err = pgxPool.Ping(context.Background()); err != nil {
		return nil, err
	}

	return pgxPool, nil
}

func ConfigureDatabase(t *testing.T) *pgxpool.Pool {
	t.Helper()

	postgresContainer, err := SetupContainer()

	t.Cleanup(func() { postgresContainer.Terminate(context.Background()) })

	assert.NoError(t, err, "запуск тестового контейнера с БД, закончился ошибкой")

	connString, err := DatabaseURL(postgresContainer)

	assert.NoError(t, err, "получение строки подключения к БД, закончилось ошибкой")

	err = RunMigrations(connString, migrationsPath)

	assert.NoError(t, err, "запуск миграций БД, закончился ошибкой")

	pgxPool, err := ConnectToDB(connString)

	t.Cleanup(func() { pgxPool.Close() })

	assert.NoError(t, err, "ошибка при попытке подключения к БД")

	return pgxPool

}

func TestUserStorage_TrackLink(t *testing.T) {
	pgxPool := ConfigureDatabase(t)
	userRepo := cleansql.NewStore(&sql.DBConfig{}, pgxPool)

	type TestCase struct {
		name        string
		userID      int64
		link        string
		expectedErr bool
		addTime     time.Time
	}

	tests := []TestCase{
		{
			name:    "Добавление первой ссылки для пользователя с id = 1",
			userID:  firstID,
			link:    "https://github.com/orlov4919/test",
			addTime: time.Now(),
		},
		{
			name:    "Добавление второй ссылки для пользователя с id = 1",
			userID:  firstID,
			link:    "https://github.com/orlov4919/test2",
			addTime: time.Now(),
		},
		{
			name:    "Добавление первой ссылки для пользователя с id = 2, которая уже есть в БД",
			userID:  secondID,
			link:    "https://github.com/orlov4919/test",
			addTime: time.Now(),
		},
		{
			name:        "Пытаемся добавить уже отслеживаемую ссылку, ожидая ошибку",
			userID:      secondID,
			link:        "https://github.com/orlov4919/test",
			addTime:     time.Now(),
			expectedErr: true,
		},
	}

	var userID int64
	var link string

	expectedRows := 1

	for _, test := range tests {
		err := userRepo.TrackLink(test.userID, test.link, test.addTime)

		if test.expectedErr {
			assert.Error(t, err, fmt.Sprintf("ошибка в тесте %s при добавлении дубля ссылки", test.name))
		} else {
			assert.NoError(t, err, fmt.Sprintf("ошибка в тесте %s при добавлении данных в БД", test.name))
		}

		rows, err := pgxPool.Query(context.Background(),
			`SELECT u.user_id, l.link_url FROM
               links AS l JOIN  userLinks as u 
               ON l.link_id = u.link_id 
               WHERE user_id = ($1) AND l.link_url = ($2)`, test.userID, test.link)

		assert.NoError(t, err, fmt.Sprintf("ошибка в тесте %s при проверке данных в БД", test.name))

		defer rows.Close()

		countRows := 0

		for rows.Next() {
			countRows++

			err = rows.Scan(&userID, &link)
			assert.NoError(t, err, fmt.Sprintf("ошибка в тесте %s при сканировании данных из БД", test.name))
		}

		assert.NoError(t, rows.Err(), fmt.Sprintf("ошибка в тесте %s при сканировании данных из БД", test.name))

		assert.Equal(t, expectedRows, countRows)
		assert.Equal(t, test.userID, userID)
		assert.Equal(t, test.link, link)
	}
}

func TestUserStorage_ChangeLastCheckTime(t *testing.T) {
	pgxPool := ConfigureDatabase(t)
	userRepo := cleansql.NewStore(&sql.DBConfig{}, pgxPool)

	type TestData struct {
		userID int64
		link   string
	}

	dataToDB := []TestData{
		{
			userID: firstID,
			link:   githubLink,
		},
		{
			userID: secondID,
			link:   stackoverflowLink,
		},
	}

	for _, data := range dataToDB {
		err := userRepo.TrackLink(data.userID, data.link, time.Now())

		assert.NoError(t, err, "ошибка при подготовке тестовых данных")
	}

	type TestCase struct {
		link    string
		newTime time.Time
	}

	tests := []TestCase{
		{
			link:    githubLink,
			newTime: time.Now().Truncate(time.Second).UTC(),
		},
		{
			link:    stackoverflowLink,
			newTime: time.Now().Truncate(time.Second).UTC(),
		},
	}

	updateTime := time.Time{}
	expectedRows := 1

	for _, test := range tests {
		err := userRepo.ChangeLastCheckTime(test.link, test.newTime)

		assert.NoError(t, err)

		rows, err := pgxPool.Query(context.Background(),
			`SELECT last_update_check FROM links
				 WHERE link_url = ($1)`, test.link)

		assert.NoError(t, err)

		defer rows.Close()

		countRows := 0

		for rows.Next() {
			countRows++

			err = rows.Scan(&updateTime)

			assert.NoError(t, err, "ошибка при сканировании времени обновления из БД")
		}

		assert.NoError(t, rows.Err())
		assert.Equal(t, expectedRows, countRows)
		assert.Equal(t, test.newTime, updateTime)
	}
}

func TestUserStorage_AllUserLinks(t *testing.T) {
	pgxPool := ConfigureDatabase(t)
	userRepo := cleansql.NewStore(&sql.DBConfig{}, pgxPool)

	type TestData struct {
		userID int64
		link   string
	}

	dataToDB := []TestData{
		{
			userID: firstID,
			link:   githubLink,
		},
		{
			userID: secondID,
			link:   stackoverflowLink,
		},
		{
			userID: secondID,
			link:   githubLink,
		},
	}

	for _, data := range dataToDB {
		err := userRepo.TrackLink(data.userID, data.link, time.Now())

		assert.NoError(t, err, "ошибка при подготовке тестовых данных")
	}

	type TestCase struct {
		userID        int64
		expectedLinks []Link
	}

	tests := []TestCase{
		{
			userID:        secondID,
			expectedLinks: []Link{stackoverflowLink, githubLink},
		},
		{
			userID:        firstID,
			expectedLinks: []Link{githubLink},
		},
	}

	for _, test := range tests {
		links, err := userRepo.AllUserLinks(test.userID)

		assert.NoError(t, err)
		assert.ElementsMatch(t, test.expectedLinks, links)
	}
}

func TestUserStorage_UserTrackLink(t *testing.T) {
	pgxPool := ConfigureDatabase(t)
	userRepo := cleansql.NewStore(&sql.DBConfig{}, pgxPool)

	type TestData struct {
		userID int64
		link   string
	}

	dataToDB := []TestData{
		{
			userID: firstID,
			link:   githubLink,
		},
		{
			userID: secondID,
			link:   stackoverflowLink,
		},
	}

	for _, data := range dataToDB {
		err := userRepo.TrackLink(data.userID, data.link, time.Now())

		assert.NoError(t, err, "ошибка при подготовке тестовых данных")
	}

	type TestCase struct {
		userID    int64
		link      Link
		trackLink bool
	}

	tests := []TestCase{
		{
			userID:    firstID,
			link:      githubLink,
			trackLink: true,
		},
		{
			userID:    secondID,
			link:      stackoverflowLink,
			trackLink: true,
		},
		{
			userID:    secondID,
			link:      githubLink,
			trackLink: false,
		},
	}

	for _, test := range tests {
		track, err := userRepo.UserTrackLink(test.userID, test.link)

		assert.NoError(t, err)
		assert.Equal(t, test.trackLink, track)
	}
}

func TestUserStorage_UserExist(t *testing.T) {
	pgxPool := ConfigureDatabase(t)
	userRepo := cleansql.NewStore(&sql.DBConfig{}, pgxPool)

	type TestData struct {
		userID int64
		link   string
	}

	dataToDB := []TestData{
		{
			userID: firstID,
			link:   githubLink,
		},
		{
			userID: secondID,
			link:   stackoverflowLink,
		},
	}

	for _, data := range dataToDB {
		err := userRepo.TrackLink(data.userID, data.link, time.Now())

		assert.NoError(t, err, "ошибка при подготовке тестовых данных")
	}

	type TestCase struct {
		userID    int64
		userExist bool
	}

	tests := []TestCase{
		{
			userID:    firstID,
			userExist: true,
		},
		{
			userID:    secondID,
			userExist: true,
		},
		{
			userID:    thirdID,
			userExist: false,
		},
	}

	for _, test := range tests {
		userExist, err := userRepo.UserExist(test.userID)

		assert.NoError(t, err)
		assert.Equal(t, test.userExist, userExist)
	}
}

func TestUserStorage_RegUser(t *testing.T) {
	pgxPool := ConfigureDatabase(t)
	userRepo := cleansql.NewStore(&sql.DBConfig{}, pgxPool)

	type TestCase struct {
		userID int64
	}

	tests := []TestCase{
		{
			userID: firstID,
		},
		{
			userID: secondID,
		},
	}

	for _, test := range tests {
		err := userRepo.RegUser(test.userID)

		assert.NoError(t, err)

		userExist, err := userRepo.UserExist(test.userID)

		assert.Equal(t, true, userExist)
	}
}

func TestUserStorage_DeleteUser(t *testing.T) {
	pgxPool := ConfigureDatabase(t)
	userRepo := cleansql.NewStore(&sql.DBConfig{}, pgxPool)

	type TestData struct {
		userID int64
		link   string
	}

	dataToDB := []TestData{
		{
			userID: firstID,
			link:   githubLink,
		},
		{
			userID: secondID,
			link:   stackoverflowLink,
		},
		{
			userID: secondID,
			link:   githubLink,
		},
	}

	for _, data := range dataToDB {
		err := userRepo.TrackLink(data.userID, data.link, time.Now())

		assert.NoError(t, err, "ошибка при подготовке тестовых данных")
	}

	type TestCase struct {
		userID int64
	}

	tests := []TestCase{
		{
			userID: firstID,
		},
		{
			userID: secondID,
		},
		{
			userID: thirdID,
		},
	}

	expectedRows := 0
	countRowsWithUser := 0

	for _, test := range tests {
		err := userRepo.DeleteUser(test.userID)

		assert.NoError(t, err)

		err = pgxPool.QueryRow(context.Background(), "SELECT COUNT(*) FROM userlinks WHERE user_id = ($1)", test.userID).
			Scan(&countRowsWithUser)

		assert.NoError(t, err, "при подсчете отслеживаемых ссылок пользователя произошла ошибка")
		assert.Equal(t, expectedRows, countRowsWithUser)

		err = pgxPool.QueryRow(context.Background(), "SELECT COUNT(*) FROM users WHERE user_id = ($1)", test.userID).
			Scan(&countRowsWithUser)

		assert.NoError(t, err, "при подсчете вхождений пользователя в таблицу users, произошла ошибка")
		assert.Equal(t, expectedRows, countRowsWithUser)
	}
}

func TestUserStorage_UntrackLink(t *testing.T) {
	pgxPool := ConfigureDatabase(t)
	userRepo := cleansql.NewStore(&sql.DBConfig{}, pgxPool)

	type TestData struct {
		userID int64
		link   string
	}

	dataToDB := []TestData{
		{
			userID: firstID,
			link:   githubLink,
		},
		{
			userID: secondID,
			link:   stackoverflowLink,
		},
		{
			userID: secondID,
			link:   githubLink,
		},
	}

	for _, data := range dataToDB {
		err := userRepo.TrackLink(data.userID, data.link, time.Now())

		assert.NoError(t, err, "ошибка при подготовке тестовых данных")
	}

	type TestCase struct {
		userID                 int64
		deletedLink            string
		trackedLinkAfterDelete int
	}

	tests := []TestCase{
		{
			userID:                 firstID,
			deletedLink:            githubLink,
			trackedLinkAfterDelete: 0,
		},
		{
			userID:                 secondID,
			deletedLink:            stackoverflowLink,
			trackedLinkAfterDelete: 1,
		},
	}

	currentCount := 0

	for _, test := range tests {
		err := userRepo.UntrackLink(test.userID, test.deletedLink)

		assert.NoError(t, err)

		err = pgxPool.QueryRow(context.Background(),
			`SELECT COUNT(*) FROM userLinks WHERE user_id = ($1) `,
			test.userID).Scan(&currentCount)

		assert.Equal(t, test.trackedLinkAfterDelete, currentCount)
	}
}
