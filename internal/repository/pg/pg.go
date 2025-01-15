package pg

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/Painkiller675/url_shortener_6750/internal/lib/merrors"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	//_ github.com/lib/pq

	"time"
)

type Storage struct {
	conn *sql.DB
	// TODO mb use logger here
}

func NewStorage(conStr string) (*Storage, error) { // TODO: mb delete error? leave only panic
	//connect to the database
	conn, err := sql.Open("pgx", conStr)
	if err != nil {
		panic(err)
	}
	//defer conn.Close()
	return &Storage{conn: conn}, nil
}

// TODO: [4 MENTOR] mb I should somehow move Bootstrap to NewStorage?
func (s *Storage) Bootstrap(ctx context.Context) error {
	// запускаем транзакцию
	tx, err := s.conn.BeginTx(ctx, nil)
	if err != nil {
		fmt.Println("[ERROR] cannot begin transaction")
		return err
	}

	// в случае неуспешного коммита все изменения транзакции будут отменены
	defer tx.Rollback()

	// создаём таблицу
	tx.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS url(
		id SERIAL PRIMARY KEY,
		alias TEXT NOT NULL UNIQUE,
		url TEXT NOT NULL);`)
	tx.ExecContext(ctx, `CREATE INDEX IF NOT EXISTS idx_alias ON url(alias);`)

	// коммитим транзакцию
	return tx.Commit()
}

func (s *Storage) StoreAlURL(ctx context.Context, alias string, url string) (int64, error) {
	const op = "postgreSQL.StoreAlURL"
	stmt, err := s.conn.Prepare("INSERT INTO url (alias, url) VALUES ($1,$2);")
	if err != nil {
		err = fmt.Errorf("%s: %w", op, err)
		return 0, err
	}

	_, err = stmt.ExecContext(ctx, alias, url) // _ = res (to ge LastId)
	if err != nil {

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
			err = merrors.ErrURLOrAliasExists
		}

		return 0, fmt.Errorf("%s: %w", op, err)
	}
	// TODO [Mentor]: LastInsertId is not supported by this driver, should I use 0 or change driver?
	/*id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	*/
	return 0, nil
}

func (s *Storage) GetOrURLByAl(ctx context.Context, alias string) (string, error) {
	const op = "postgreSQL.GetOrURLByAl"
	row := s.conn.QueryRowContext(ctx, "SELECT url FROM url WHERE alias=$1;", alias)
	var orURL string
	err := row.Scan(&orURL)
	if err != nil {
		// if URL doesn't exist
		if errors.Is(err, pgx.ErrNoRows) { // TODO [4 MENTOR]: why it doesn't work?!
			return "", fmt.Errorf("%s: %w", op, merrors.ErrURLNotFound)
		}
		// other possible errors
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return orURL, nil

}

func (s *Storage) Ping(ctx context.Context) error {
	fmt.Println("[INFO] ping from the pg")
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()
	if err := s.conn.PingContext(ctx); err != nil {
		return err
	}
	// connection is established
	return nil

}
