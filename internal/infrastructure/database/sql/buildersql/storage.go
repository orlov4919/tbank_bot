package buildersql

import (
	"context"
	"fmt"
	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	"github.com/jackc/pgx/v5"
	"linkTraccer/internal/application/scrapper/scrapservice"
	"linkTraccer/internal/domain/scrapper"
	"linkTraccer/internal/infrastructure/database/sql"
	"time"
)

const (
	usersCap = 20
	linkCap  = 20
)

type Link = scrapper.Link
type User = scrapper.User
type LinkInfo = scrapper.LinkInfo
type LinkID = scrapper.LinkID
type Tag = scrapper.Tag

type UserStorage struct {
	dsn       string
	batchSize int
	db        *pgx.Conn
}

type linkPaginator struct {
	db         *pgx.Conn
	lastLinkID int64
	limit      int
}

func NewStore(cfg *sql.DBConfig) *UserStorage {
	return &UserStorage{dsn: cfg.ToDSN(), batchSize: cfg.BatchSize}
}

func (u *UserStorage) Open() error {
	con, err := pgx.Connect(context.Background(), u.dsn)

	fmt.Println(u.dsn)

	if err != nil {
		return fmt.Errorf("при создании conn возникла ошибка: %w", err)
	}

	if err = con.Ping(context.Background()); err != nil {
		return fmt.Errorf("при установлении конекта с БД произошла ошибка: %w", err)
	}

	u.db = con

	return nil
}

func (u *UserStorage) AllLinks() ([]LinkInfo, error) {
	var linkID LinkID
	var url Link
	var date time.Time

	sqlCmd, _, _ := goqu.From("links").Select("link_id,link_url,last_update_check").ToSQL()
	rows, err := u.db.Query(context.Background(), sqlCmd)

	if err != nil {
		return nil, fmt.Errorf("при получении всех ссылок произошла ошибка: %w", err)
	}

	defer rows.Close()

	links := make([]LinkInfo, 0, 1000)

	for rows.Next() {
		if err = rows.Scan(&linkID, &url, &date); err != nil {
			return nil, fmt.Errorf("ошибка при получении всех ссылок: %w", err)
		}
		links = append(links, LinkInfo{ID: linkID, URL: url, LastUpdate: date})
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при получении всех ссылок: %w", err)
	}

	return links, nil
}

func (u *UserStorage) TrackLink(userID User, link Link, addTime time.Time) error {
	var link_id int64

	tx, err := u.db.Begin(context.Background())

	if err != nil {
		return fmt.Errorf("при попытке создать соединение для транзакции произошла ошибка: %w", err)
	}

	sqlCmd, _, _ := goqu.Dialect("postgres").
		From("id").
		With("id", goqu.Insert("links").
			Returning("link_id").
			Cols("link_url", "last_update_check").
			Vals(goqu.Vals{goqu.L("$1"), goqu.L("$2")}).
			OnConflict(goqu.DoNothing())).Select("id.link_id").
		UnionAll(goqu.From("links").Select("link_id").Where(goqu.Ex{"link_url": goqu.L("$1")})).
		ToSQL()

	if err := tx.QueryRow(context.Background(), sqlCmd, link, addTime).
		Scan(&link_id); err != nil {
		tx.Rollback(context.Background())
		return fmt.Errorf("ошибка при добавлении в таблицу links: %w", err)
	}

	sqlCmd, _, _ = goqu.Insert("users").
		Cols("user_id").
		Vals(goqu.Vals{goqu.L("$1")}).
		OnConflict(goqu.DoNothing()).ToSQL()

	if _, err := tx.Exec(context.Background(), sqlCmd, userID); err != nil {
		tx.Rollback(context.Background())
		return fmt.Errorf("ошибка при добавлении в таблицу users")
	}

	sqlCmd, _, _ = goqu.Insert("userlinks").
		Cols("user_id", "link_id").
		Vals(goqu.Vals{goqu.L("$1"), goqu.L("$2")}).
		OnConflict(goqu.DoNothing()).ToSQL()

	if _, err := tx.Exec(context.Background(), sqlCmd, userID, link_id); err != nil {
		tx.Rollback(context.Background())
		return fmt.Errorf("ошибка при добавлении в таблицу userlinks")
	}

	if err = tx.Commit(context.Background()); err != nil {
		return fmt.Errorf("ошибка при создании коммита транзакции: %w", err)
	}

	return nil
}

func (u *UserStorage) ChangeLastCheckTime(link Link, checkTime time.Time) error {
	sqlCmd, _, _ := goqu.Update("links").
		Set(goqu.Record{"last_update_check": goqu.L("$2")}).
		Where(goqu.Ex{"link_url": goqu.L("$1")}).
		ToSQL()

	if _, err := u.db.Exec(context.Background(), sqlCmd, link, checkTime); err != nil {
		return fmt.Errorf("ошибка при изменении времени: %w", err)
	}

	return nil
}

func (u *UserStorage) UsersWhoTrackLink(linkID LinkID) ([]User, error) {
	var user User

	sqlCms, _, _ := goqu.From("userlinks").
		Select("user_id").
		Where(goqu.Ex{"link_id": goqu.L("$1")}).
		ToSQL()

	rows, err := u.db.Query(context.Background(), sqlCms, linkID)

	if err != nil {
		return nil, fmt.Errorf("ошибка при получении отслеживающих ссылку пользователей: %w", err)
	}

	users := make([]User, 0, usersCap)

	for rows.Next() {
		if err = rows.Scan(&user); err != nil {
			return nil, fmt.Errorf("ошибка при чтении строки: %w", err)
		}

		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при получении всех ссылок: %w", err)
	}

	return users, nil
}

func (u *UserStorage) AllUserLinks(userID User) ([]Link, error) {
	var link Link

	sqlCmd, _, _ := goqu.From("links").
		Select("link_url").
		Join(goqu.T("userlinks"), goqu.On(goqu.Ex{"links.link_id": goqu.I("userlinks.link_id")})).
		Where(goqu.Ex{"userlinks.user_id": goqu.L("$1")}).
		ToSQL()

	rows, err := u.db.Query(context.Background(), sqlCmd, userID)

	if err != nil {
		return nil, fmt.Errorf("ошибка при выполнении запроса на получение всех ссылок пользователя: %w", err)
	}

	links := make([]Link, 0, linkCap)

	for rows.Next() {
		if err = rows.Scan(&link); err != nil {
			return nil, fmt.Errorf("ошибка при чтении строки: %w", err)
		}

		links = append(links, link)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при получении всех ссылок пользователя: %w", err)
	}

	return links, nil
}

func (u *UserStorage) UserTrackLink(userID User, URL Link) (bool, error) { // переписать на Query Row
	var link Link

	sqlCmd, _, _ := goqu.From("links").
		Select("link_url").
		Join(goqu.T("userlinks"), goqu.On(goqu.Ex{"links.link_id": goqu.I("userlinks.link_id")})).
		Where(goqu.Ex{"userlinks.user_id": goqu.L("$1")}, goqu.Ex{"links.link_url": goqu.L("$2")}).
		ToSQL()

	rows, err := u.db.Query(context.Background(), sqlCmd, userID, URL)

	if err != nil {
		return false, fmt.Errorf("ошибка при выполнении запроса на получение всех ссылок пользователя: %w", err)
	}

	for rows.Next() {
		if err = rows.Scan(&link); err != nil {
			return false, fmt.Errorf("ошибка при чтении строки: %w", err)
		}
	}

	if err = rows.Err(); err != nil {
		return false, fmt.Errorf("ошибка при получении всех ссылок пользователя: %w", err)
	}

	return link != "", nil
}

func (u *UserStorage) UserExist(UserID User) (bool, error) {
	var user User

	sqlCmd, _, _ := goqu.From("users").
		Select("user_id").
		Where(goqu.Ex{"user_id": goqu.L("$1")}).
		ToSQL()

	rows, err := u.db.Query(context.Background(), sqlCmd, UserID)

	if err != nil {
		return false, fmt.Errorf("ошибка при создании соединения:%w", err)
	}

	defer rows.Close()

	for rows.Next() {
		if err = rows.Scan(&user); err != nil {
			return false, fmt.Errorf("ошибка при чтении строки: %w", err)
		}
	}

	if err = rows.Err(); err != nil {
		return false, fmt.Errorf("ошибка при получении всех ссылок пользователя: %w", err)
	}

	return user == UserID, nil
}

func (u *UserStorage) RegUser(UserID User) error {
	sqlCmd, _, _ := goqu.Insert("users").
		Cols("user_id").
		Vals(goqu.Vals{goqu.L("$1")}).
		OnConflict(goqu.DoNothing()).
		ToSQL()

	if _, err := u.db.Exec(context.Background(), sqlCmd, UserID); err != nil {
		return fmt.Errorf("ошибка при добавлении нового пользователя: %w", err)
	}

	return nil
}

// проработать этот метод на все ссылки которые отслеживал только он

func (u *UserStorage) DeleteUser(user User) error {
	sqlCmd, _, _ := goqu.Delete("users").Where(goqu.Ex{"user_id": goqu.L("$1")}).ToSQL()

	if _, err := u.db.Exec(context.Background(), sqlCmd, user); err != nil {
		return fmt.Errorf("ошибка при удалении пользователя: %w", err)
	}

	return nil
}

// доработать этот метод, если не один пользователь не отслеживает ссылку

func (u *UserStorage) UntrackLink(user User, link Link) error {

	if _, err := u.db.Exec(context.Background(), "DELETE FROM userlinks USING links WHERE userlinks.link_id = links.link_id AND links.link_url = ($1) AND userlinks.user_id = ($2)", link, user); err != nil {
		return fmt.Errorf("ошибка во время удаления ссылки пользователя: %w", err)
	}

	return nil
}

func (u *UserStorage) Close() error {
	return u.db.Close(context.Background())
}

func (u *UserStorage) NewLinksPaginator() scrapservice.LinkPaginator {
	return &linkPaginator{db: u.db, limit: u.batchSize}
}

func (l *linkPaginator) LinksBatch() ([]LinkInfo, error) {
	var id int64

	var link string

	var lastCheck time.Time

	var linkInfo LinkInfo

	// впихнуть where по тому кто не обновлялся более 5 минут к примеру

	sqlCmd, _, _ := goqu.From("links").
		Select("link_id", "link_url", "last_update_check").
		Where(goqu.C("link_id").Gt(goqu.L("$1")),
			goqu.C("last_update_check").Lt(goqu.L("CURRENT_TIMESTAMP - INTERVAL '5 minutes'"))).
		Order(goqu.I("link_id").Asc()).Limit(uint(l.limit)).
		ToSQL()

	rows, err := l.db.Query(context.Background(), sqlCmd, l.lastLinkID)

	if err != nil {
		return nil, fmt.Errorf("ошика при выполнении запроса на получение пачки ссылок: %w", err)
	}

	defer rows.Close()

	links := make([]LinkInfo, 0, l.limit)

	for rows.Next() {
		if err = rows.Scan(&id, &link, &lastCheck); err != nil {
			return nil, fmt.Errorf("ошика при сканировании ссылок: %w", err)
		}

		linkInfo.ID = id
		linkInfo.URL = link
		linkInfo.LastUpdate = lastCheck

		links = append(links, linkInfo)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("ошика при сканировании ссылок: %w", err)
	}

	l.lastLinkID = id

	return links, nil
}
