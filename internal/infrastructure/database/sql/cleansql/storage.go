package cleansql

import (
	"context"
	"fmt"
	"linkTraccer/internal/application/scrapper/scrapservice"
	"linkTraccer/internal/domain/scrapper"
	"linkTraccer/internal/infrastructure/database/sql"
	"linkTraccer/internal/infrastructure/database/sql/transactor"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	usersCap = 20
	linkCap  = 20
)

type LinkInfo = scrapper.LinkInfo
type LinkID = scrapper.LinkID
type Tag = scrapper.Tag

type UserStorage struct {
	batchSize uint
	db        *pgxpool.Pool
}

func NewStore(dbConfig *sql.DBConfig, pgxPool *pgxpool.Pool) *UserStorage {
	return &UserStorage{db: pgxPool, batchSize: dbConfig.BatchSize}
}

// возможно в этой части кода нужно выделение отдельного соединения, для того что бы метод всегда выполнялся транзакционно

func (u *UserStorage) TrackLink(ctx context.Context, userID scrapper.User, link scrapper.Link, addTime time.Time) error {
	var linkID int64

	conn := transactor.GetQuerier(ctx, u.db)

	err := conn.QueryRow(context.Background(), `WITH id AS (INSERT INTO links(link_url,last_update_check) values 
 														($1,$2) ON CONFLICT (link_url) DO NOTHING RETURNING link_id)
														SELECT id.link_id FROM id UNION ALL SELECT link_id FROM links
														WHERE link_url = ($1);`, link, addTime).Scan(&linkID)

	if err != nil {
		return fmt.Errorf("ошибка при добавлении в таблицу links: %w", err)
	}

	_, err = conn.Exec(context.Background(),
		`INSERT INTO users(user_id) values ($1) ON CONFLICT (user_id) DO NOTHING`, userID)

	if err != nil {
		return fmt.Errorf("ошибка при добавлении в таблицу users")
	}

	if _, err := conn.Exec(context.Background(), `INSERT INTO userlinks(user_id,link_id) values ($1,$2)`,
		userID, linkID); err != nil {
		return fmt.Errorf("ошибка при добавлении в таблицу userlinks")
	}

	return nil
}

func (u *UserStorage) ChangeLastCheckTime(link scrapper.Link, checkTime time.Time) error {
	_, err := u.db.Exec(context.Background(), "UPDATE links SET last_update_check=($2) WHERE link_url = ($1)",
		link, checkTime)

	if err != nil {
		return fmt.Errorf("ошибка при изменении времени обновлениия ссылки: %w", err)
	}

	return nil
}

func (u *UserStorage) UsersWhoTrackLink(linkID LinkID) ([]scrapper.User, error) {
	var user scrapper.User

	rows, err := u.db.Query(context.Background(), "SELECT user_id FROM userlinks WHERE link_id = ($1) ", linkID)

	if err != nil {
		return nil, fmt.Errorf("ошибка при получении отслеживающих ссылку пользователей: %w", err)
	}

	users := make([]scrapper.User, 0, usersCap)

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

func (u *UserStorage) AllUserLinks(userID scrapper.User) ([]scrapper.Link, error) {
	var link scrapper.Link

	rows, err := u.db.Query(context.Background(),
		`SELECT link_url FROM links 
    		 JOIN userlinks ON links.link_id = userlinks.link_id 
             WHERE userlinks.user_id = ($1)`, userID)

	if err != nil {
		return nil, fmt.Errorf("ошибка при выполнении запроса на получение всех ссылок пользователя: %w", err)
	}

	links := make([]scrapper.Link, 0, linkCap)

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

func (u *UserStorage) UserTrackLink(userID scrapper.User, url scrapper.Link) (bool, error) {
	var link scrapper.Link

	err := u.db.QueryRow(context.Background(),
		`SELECT link_url FROM links JOIN userLinks ON links.link_id = userLinks.link_id
             WHERE userLinks.user_id = ($1) AND links.link_url = ($2)`, userID, url).Scan(&link)

	if err == pgx.ErrNoRows {
		return false, nil
	}

	if err != nil {
		return false, fmt.Errorf("ошибка при проверки ссылки пользователя: %w", err)
	}

	return true, nil
}

func (u *UserStorage) UserExist(userID scrapper.User) (bool, error) {
	var user scrapper.User

	err := u.db.QueryRow(context.Background(),
		"SELECT user_id FROM users WHERE user_id = ($1)", userID).Scan(&user)

	if err == pgx.ErrNoRows {
		return false, nil
	}

	if err != nil {
		return false, fmt.Errorf("ошибка при проверки пользователя в БД: %w", err)
	}

	return user == userID, nil
}

func (u *UserStorage) RegUser(userID scrapper.User) error {
	_, err := u.db.Exec(context.Background(),
		"INSERT INTO users(user_id) VALUES ($1) ON CONFLICT (user_id) DO NOTHING", userID)

	if err != nil {
		return fmt.Errorf("ошибка при добавлении нового пользователя: %w", err)
	}

	return nil
}

func (u *UserStorage) DeleteUser(ctx context.Context, user scrapper.User) (err error) {
	conn := transactor.GetQuerier(ctx, u.db)

	if _, err = conn.Exec(context.Background(), "DELETE FROM userlinks WHERE user_id = ($1);", user); err != nil {
		return fmt.Errorf("ошибка при удалении пользователя из таблицы userlinks: %w", err)
	}

	if _, err = conn.Exec(context.Background(), "DELETE FROM users WHERE user_id = ($1);", user); err != nil {
		return fmt.Errorf("ошибка при удалении пользователя из таблицы users: %w", err)
	}

	return nil
}

func (u *UserStorage) UntrackLink(user scrapper.User, link scrapper.Link) error {
	_, err := u.db.Exec(context.Background(),
		`DELETE FROM userlinks USING links 
             WHERE userlinks.link_id = links.link_id AND links.link_url = ($1) AND userlinks.user_id = ($2)`,
		link, user)

	if err != nil {
		return fmt.Errorf("ошибка во время удаления ссылки пользователя: %w", err)
	}

	return nil
}

func (u *UserStorage) DeleteUntrackedLinks() error {
	_, err := u.db.Exec(context.Background(), `DELETE FROM links 
       									 WHERE link_id NOT IN (SELECT  link_id FROM userlinks)`)

	if err != nil {
		return fmt.Errorf("ошибка при удалении не отслеживаемых ссылок: %w", err)
	}

	return nil
}

type linkPaginator struct {
	db         *pgxpool.Pool
	lastLinkID int64
	limit      uint
	hasLinks   bool
}

func (u *UserStorage) NewLinksPaginator() scrapservice.LinkPaginator {
	return &linkPaginator{db: u.db, limit: u.batchSize, hasLinks: true}
}

func (l *linkPaginator) HasLinks() bool {
	return l.hasLinks
}

func (l *linkPaginator) LinksBatch() ([]*LinkInfo, error) {
	var id int64

	var link string

	var lastCheck time.Time

	rows, err := l.db.Query(context.Background(),
		`SELECT link_id, link_url, last_update_check FROM links 
             WHERE link_id > ($1) AND CURRENT_TIMESTAMP - last_update_check > '5 minutes' 
             ORDER BY link_id ASC LIMIT ($2);`, l.lastLinkID, l.limit)

	if err != nil {
		return nil, fmt.Errorf("ошика при выполнении запроса на получение пачки ссылок: %w", err)
	}

	defer rows.Close()

	links := make([]*LinkInfo, 0, l.limit)

	for rows.Next() {
		linkInfo := &LinkInfo{}

		if err = rows.Scan(&id, &link, &lastCheck); err != nil {
			return nil, fmt.Errorf("ошика при сканировании ссылок: %w", err)
		}

		linkInfo.ID = id
		linkInfo.URL = link
		linkInfo.LastUpdate = lastCheck

		links = append(links, linkInfo)
	}

	l.lastLinkID = id

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("ошика при сканировании ссылок: %w", err)
	}

	if len(links) == 0 {
		l.hasLinks = false
	}

	return links, nil
}
