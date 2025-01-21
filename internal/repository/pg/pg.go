package pg

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/Painkiller675/url_shortener_6750/internal/lib/merrors"
	"github.com/Painkiller675/url_shortener_6750/internal/models"
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
		url TEXT NOT NULL UNIQUE);`)
	tx.ExecContext(ctx, `CREATE INDEX IF NOT EXISTS idx_alias ON url(alias);`)

	// коммитим транзакцию
	return tx.Commit()
}

func (s *Storage) StoreAlURL(ctx context.Context, alias string, url string) (int64, error) {
	const op = "pg.StoreAlURL"
	stmt, err := s.conn.Prepare("INSERT INTO url (alias, url) VALUES ($1,$2);")
	if err != nil {
		err = fmt.Errorf("%s: %w", op, err)
		return 0, err
	}
	fmt.Println("1. alias = ", alias)
	_, err = stmt.ExecContext(ctx, alias, url) // _ = res (to ge LastId)
	if err != nil {
		fmt.Printf("%s: %s\n", op, err)
		// TODO: what does it actually mean???
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
			fmt.Println("if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code)")
			if pgErr.Code == pgerrcode.UniqueViolation { // if we have the try to short existed url
				fmt.Println("if pgErr.Code == pgerrcode.UniqueViolation ")
				fmt.Println("2. alias = ", alias)
				//existedAlias, err := s.GetAlByURL(ctx, url)
				//if err != nil {
				//	return 0, err
				//}
				err = merrors.ErrURLOrAliasExists // TODO [MENTOR]: is it OK way ?
				return 0, err
				//return 0, &models.ExistsURLError{ // TODO [MENTOR] Delete the model
				//	ExistedAlias: existedAlias,
				//	Err:          merrors.ErrURLOrAliasExists,
			}
		}
		//err = merrors.ErrURLOrAliasExists
	}
	return 0, nil
	//return 0, fmt.Errorf("%s: %w", op, err)
}

// TODO [Mentor]: LastInsertId is not supported by this driver, should I use 0 or change the driver?
/*id, err := res.LastInsertId()
if err != nil {
	return 0, fmt.Errorf("%s: %w", op, err)
}
*/

func (s *Storage) GetOrURLByAl(ctx context.Context, alias string) (string, error) {
	const op = "postgreSQL.GetOrURLByAl"
	// TODO [MENTOR] mb I should put all the queries into the constants?!
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

func (s *Storage) GetAlByURL(ctx context.Context, url string) (string, error) {
	const op = "postgreSQL.GetOrURLByAl"
	// TODO [MENTOR] mb I should put all the queries into the constants?!
	row := s.conn.QueryRowContext(ctx, "SELECT alias FROM url WHERE url=$1;", url)
	var alias string
	err := row.Scan(&alias)
	fmt.Println("REPEAT FROM PG, alias =", alias)
	if err != nil {
		// if alias doesn't exist // TODO: [MENTOR] it's impossible, should I del that?
		if errors.Is(err, pgx.ErrNoRows) { // TODO [4 MENTOR]: why it doesn't work?!
			return "", fmt.Errorf("%s: %w", op, merrors.ErrURLOrAliasExists)
		}
		// other possible errors
		fmt.Println("other possible errors ")
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return alias, nil

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

func (s *Storage) SaveBatchURL(ctx context.Context, corURLSh *[]models.JSONBatStructIDOrSh) (*[]models.JSONBatStructToSerResp, error) {
	const op = "pg.SaveBatchURL"
	// запускаем транзакцию
	tx, err := s.conn.BeginTx(ctx, nil)
	if err != nil {
		fmt.Println("[ERROR] cannot begin transaction")
		return nil, err
	}
	// в случае неуспешного коммита все изменения транзакции будут отменены
	defer tx.Rollback()
	// create value to store data for response
	toResp := make([]models.JSONBatStructToSerResp, 0) // TODO [MENTOR]: is it ok allocation? why len(*corURLSh) is false instead of 0?
	// fill transaction with insert queries:

	for _, idURLSh := range *corURLSh {
		_, err := s.StoreAlURL(ctx, idURLSh.ShortURL, idURLSh.OriginalURL)
		// TODO: if exists => get it from DB for response
		if err != nil {
			if errors.Is(err, merrors.ErrURLOrAliasExists) {
				///existAl, err := s.GetAlByURL(ctx, idURLSh.ShortURL)
				///if err != nil {
				///	return nil, fmt.Errorf("%s: %w", op, err)
				//}
				//filling for response
				toResp = append(toResp, models.JSONBatStructToSerResp{
					CorrelationID: idURLSh.CorrelationID,
					ShortURL:      idURLSh.ShortURL, // use the same shortURL cause we use md5
				})
				continue

				// TODO: wrap all errors from StoreAlURL or use logger right here

			} else { // other possible errors
				fmt.Println("SaveBatch another error!")
				return nil, fmt.Errorf("%s: %w", op, err)
			}
		}

		toResp = append(toResp, models.JSONBatStructToSerResp{
			CorrelationID: idURLSh.CorrelationID,
			ShortURL:      idURLSh.ShortURL,
		})

	}

	// коммитим транзакцию
	err = tx.Commit()
	if err != nil {
		// TODO [MENTOR]: mb I should use logger here and in the storage ?
		return nil, err
	}

	return &toResp, nil
}
