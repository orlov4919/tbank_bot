package cleansql

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"

	//"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
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
	batchSize int
	db        *pgxpool.Pool
}

type linkPaginator struct {
	db         *pgxpool.Pool
	lastLinkID int64
	limit      int
}

func NewStore(dbConfig *sql.DBConfig, pgxPool *pgxpool.Pool) *UserStorage {
	return &UserStorage{db: pgxPool, batchSize: dbConfig.BatchSize}
}

func (u *UserStorage) AllLinks() ([]LinkInfo, error) {
	var linkID LinkID
	var url Link
	var date time.Time

	sqlCmd := "SELECT link_id,link_url,last_update_check FROM links"
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

	if err := tx.QueryRow(context.Background(), "WITH id AS (INSERT INTO links(link_url,last_update_check) values ($1,$2) ON CONFLICT (link_url) DO NOTHING RETURNING link_id) SELECT id.link_id FROM id UNION ALL SELECT link_id FROM links WHERE link_url = ($1);", link, addTime).
		Scan(&link_id); err != nil {
		tx.Rollback(context.Background())
		return fmt.Errorf("ошибка при добавлении в таблицу links: %w", err)
	}

	if _, err := tx.Exec(context.Background(), "INSERT INTO users(user_id) values ($1) ON CONFLICT (user_id) DO NOTHING", userID); err != nil {
		tx.Rollback(context.Background())
		return fmt.Errorf("ошибка при добавлении в таблицу users")
	}

	if _, err := tx.Exec(context.Background(), "INSERT INTO userlinks(user_id,link_id) values ($1,$2) ON CONFLICT (user_id,link_id) DO NOTHING", userID, link_id); err != nil {
		tx.Rollback(context.Background())
		return fmt.Errorf("ошибка при добавлении в таблицу userlinks")
	}

	if err = tx.Commit(context.Background()); err != nil {
		return fmt.Errorf("ошибка при создании коммита транзакции: %w", err)
	}

	return nil
}

func (u *UserStorage) ChangeLastCheckTime(link Link, checkTime time.Time) error {
	if _, err := u.db.Exec(context.Background(), "UPDATE links SET last_update_check=($2) WHERE link_url = ($1)", link, checkTime); err != nil {
		return fmt.Errorf("ошибка при изменении времени: %w", err)
	}

	return nil
}

func (u *UserStorage) UsersWhoTrackLink(linkID LinkID) ([]User, error) {
	var user User

	rows, err := u.db.Query(context.Background(), "SELECT user_id FROM userlinks WHERE link_id = ($1) ", linkID)

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

	rows, err := u.db.Query(context.Background(), "SELECT link_url FROM links JOIN userlinks ON links.link_id = userlinks.link_id WHERE userlinks.user_id = ($1)", userID)

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

	rows, err := u.db.Query(context.Background(), "SELECT link_url FROM links JOIN userlinks ON links.link_id = userlinks.link_id WHERE userlinks.user_id = ($1) AND links.link_url = ($2)", userID, URL)

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

	rows, err := u.db.Query(context.Background(), "SELECT user_id FROM users WHERE user_id = ($1)", UserID)

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

	if _, err := u.db.Exec(context.Background(), "INSERT INTO users(user_id) VALUES ($1) ON CONFLICT (user_id) DO NOTHING", UserID); err != nil {
		return fmt.Errorf("ошибка при добавлении нового пользователя: %w", err)
	}

	return nil
}

func (u *UserStorage) DeleteUser(user User) (err error) {
	tx, err := u.db.Begin(context.Background())

	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			err = fmt.Errorf("ошибка при выполнении транзакции удаляющей пользователя: %w", err)

			if rollbackErr := tx.Rollback(context.Background()); rollbackErr != nil {
				err = errors.Join(err, rollbackErr)
			}
		}
	}()

	fmt.Println(user)

	if _, err = tx.Exec(context.Background(), "DELETE FROM userlinks WHERE user_id = ($1);", user); err != nil {
		return err
	}

	if _, err = tx.Exec(context.Background(), "DELETE FROM users WHERE user_id = ($1);", user); err != nil {
		return fmt.Errorf("ошибка при удалении пользователя: %w", err)
	}

	if err = tx.Commit(context.Background()); err != nil {
		return err
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

func (u *UserStorage) NewLinksPaginator() scrapservice.LinkPaginator {
	return &linkPaginator{db: u.db, limit: u.batchSize}
}

func (l *linkPaginator) LinksBatch() ([]LinkInfo, error) {
	var id int64

	var link string

	var lastCheck time.Time

	var linkInfo LinkInfo

	// впихнуть where по тому кто не обновлялся более 5 минут к примеру

	rows, err := l.db.Query(context.Background(), "SELECT link_id, link_url, last_update_check FROM links WHERE link_id > ($1) AND CURRENT_TIMESTAMP - last_update_check > '5 minutes' ORDER BY link_id ASC LIMIT ($2);", l.lastLinkID, l.limit)

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
