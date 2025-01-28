package pg

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/Painkiller675/url_shortener_6750/internal/lib/merrors"
	"github.com/Painkiller675/url_shortener_6750/internal/models"
	"github.com/Painkiller675/url_shortener_6750/internal/service"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"

	//_ github.com/lib/pq

	"time"
)

type Storage struct {
	conn   *sql.DB
	logger *zap.Logger
	// TODO mb use logger here
}

func NewStorage(ctx context.Context, conStr string) (*Storage, error) { // TODO: mb delete error? leave only panic

	//connect to the database
	conn, err := sql.Open("pgx", conStr)
	if err != nil {
		return nil, err
	}
	// запускаем транзакцию
	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		fmt.Println("[ERROR] cannot begin transaction")
		return nil, err
	}
	// в случае неуспешного коммита все изменения транзакции будут отменены
	defer tx.Rollback()
	/*defer func() {
		if err := tx.Rollback(); err != nil {
			fmt.Println("[ERROR] cannot rollback transaction") // TODO [MENTOR]: Why it panics ???
		}
	}()*/

	// создаём таблицу
	tx.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS url(
		id SERIAL NOT NULL PRIMARY KEY,
		alias TEXT NOT NULL UNIQUE,
		url TEXT NOT NULL UNIQUE,
    	userId TEXT NOT NULL UNIQUE,
		created TIMESTAMP with time zone NOT NULL DEFAULT now());`)
	tx.ExecContext(ctx, `CREATE INDEX IF NOT EXISTS idx_alias ON url(alias);`)

	// коммитим транзакцию
	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	//defer conn.Close()
	return &Storage{conn: conn}, nil
}

func (s *Storage) StoreAlURL(ctx context.Context, alias string, url string, userID string) (int64, error) {
	const op = "pg.StoreAlURL"
	stmt, err := s.conn.Prepare("INSERT INTO url (alias, url, userId) VALUES ($1,$2, $3);")
	if err != nil {
		err = fmt.Errorf("%s: %w", op, err)
		return 0, err
	}
	fmt.Println("1. alias = ", alias)
	_, err = stmt.ExecContext(ctx, alias, url, userID) // _ = res (to ge LastId)
	if err != nil {
		fmt.Printf("%s: %s\n", op, err)
		// TODO: what does it actually mean???
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
			if pgErr.Code == pgerrcode.UniqueViolation { // if we have the try to short existed url
				err = merrors.ErrURLOrAliasExists // TODO [MENTOR]: is it OK way ?
				return 0, err

			}
		}

	}
	return 0, nil

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

func (s *Storage) GetDataByUserID(ctx context.Context, userID string) (*[]models.UserURLS, error) {
	const op = "pg.GetDataByUserID"
	rows, err := s.conn.QueryContext(ctx, "SELECT alias, url FROM url WHERE userId=$1;", userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) { // no rows for  a specific user
			return nil, merrors.ErrURLNotFound
		} // TODO: CHECK THAT!!!!!!!!!!!!!!!!!1
		// other possible errors
		return nil, fmt.Errorf("failed to get user URLs [%s]: %w", op, err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			s.logger.Error("failed to close rows", zap.String("place", op), zap.Error(err))
		}
	}()
	defer func() {
		if err := rows.Err(); err != nil {
			s.logger.Error("rows.Err() issue", zap.String("place", op), zap.Error(err))
		}
	}()
	var userData []models.UserURLS

	for rows.Next() {
		var usrData models.UserURLS
		if err := rows.Scan(&usrData.ShortURL, &usrData.OriginalURL); err != nil {
			return nil, fmt.Errorf("can's scan the row [%s]: %w", op, err)
		}
		userData = append(userData, usrData)
	}

	return &userData, nil
}

func (s *Storage) GetAlByURL(ctx context.Context, url string) (string, error) {
	const op = "postgreSQL.GetOrURLByURL"
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
	defer func() { // TODO [MENTOR] : mb somehow use zap here?
		if err := tx.Rollback(); err != nil {
			fmt.Println("[ERROR] rollback error!")
		}
	}()
	// create value to store data for response
	toResp := make([]models.JSONBatStructToSerResp, 0) // TODO [MENTOR]: is it ok allocation? why len(*corURLSh) is false instead of 0?
	// fill transaction with insert queries:

	for _, idURLSh := range *corURLSh {
		_, err := s.StoreAlURL(ctx, idURLSh.ShortURL, idURLSh.OriginalURL, service.GetRandString(time.Now().String())) //TODO: correct that?
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
