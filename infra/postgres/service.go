package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"time"
	"x-bank-ms-bank/cerrors"
	transaction_manager "x-bank-ms-bank/core/transaction-manager"
	"x-bank-ms-bank/core/web"
	"x-bank-ms-bank/ercodes"
)

type (
	Service struct {
		db *sql.DB
	}
)

func NewService(login, password, host string, port int, database string, maxCons int) (Service, error) {
	db, err := sql.Open("pgx", fmt.Sprintf("postgres://%s:%s@%s:%d/%s", login, password, host, port, database))
	if err != nil {
		return Service{}, err
	}

	db.SetMaxOpenConns(maxCons)

	if err = db.Ping(); err != nil {
		return Service{}, err
	}

	return Service{
		db: db,
	}, err
}

func (s *Service) GetUserAccounts(ctx context.Context, userId int64) ([]web.UserAccountData, error) {
	const query = `
SELECT accounts."id", "balanceCents", "status" 
FROM accounts 
    LEFT JOIN "accountOwners" ON "ownerId" = "accountOwners".id 
WHERE "userId" = $1
`

	rows, err := s.db.QueryContext(ctx, query, userId)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, s.wrapQueryError(err)
	}

	var userAccountsData []web.UserAccountData
	for rows.Next() {
		var data web.UserAccountData
		if err = rows.Scan(&data.Id, &data.BalanceCents, &data.Status); err != nil {
			return nil, s.wrapScanError(err)
		}
		userAccountsData = append(userAccountsData, data)
	}

	return userAccountsData, nil
}

func (s *Service) OpenUserAccount(ctx context.Context, userId int64) error {
	const query = `SELECT "id" FROM "accountOwners" WHERE "userId" = $1`

	row := s.db.QueryRowContext(ctx, query, userId)
	if err := row.Err(); err != nil {
		return s.wrapQueryError(err)
	}

	var accountOwnerId int64
	if err := row.Scan(&accountOwnerId); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			accountOwnerId, err = s.createAccountOwner(ctx, userId)
			if err != nil {
				return err
			}
		} else {
			return s.wrapScanError(err)
		}
	}

	const openAccountQuery = `INSERT INTO accounts ("ownerId") VALUES ($1)`
	_, err := s.db.ExecContext(ctx, openAccountQuery, accountOwnerId)
	if err != nil {
		return s.wrapQueryError(err)
	}
	return nil
}

func (s *Service) createAccountOwner(ctx context.Context, userId int64) (int64, error) {
	const query = `INSERT INTO "accountOwners" ("userId") VALUES ($1) RETURNING id`

	row := s.db.QueryRowContext(ctx, query, userId)
	if err := row.Err(); err != nil {
		return 0, s.wrapQueryError(err)
	}

	var id int64
	if err := row.Scan(&id); err != nil {
		return 0, s.wrapScanError(err)
	}
	return id, nil
}

func (s *Service) BlockUserAccount(ctx context.Context, accountId int64) error {
	const query = `UPDATE accounts SET status = 'BLOCKED' WHERE id = $1`

	_, err := s.db.ExecContext(ctx, query, accountId)
	if err != nil {
		return s.wrapQueryError(err)
	}
	return nil
}

func (s *Service) GetAccountHistory(ctx context.Context, accountId, limit, offset int64) ([]web.AccountTransactionsData, int64, error) {
	const query = `SELECT "senderId", "receiverId", "status", "createdAt", "amountCents", "description" FROM transactions 
					WHERE "senderId" = @accountId OR "receiverId" = @accountId ORDER BY "createdAt" DESC LIMIT @limit OFFSET @offset`

	rows, err := s.db.QueryContext(ctx, query, pgx.NamedArgs{
		"accountId": accountId,
		"limit":     limit,
		"offset":    offset,
	})
	if err != nil {
		return nil, 0, s.wrapQueryError(err)
	}

	var accountTransactionsData []web.AccountTransactionsData
	for rows.Next() {
		var data web.AccountTransactionsData
		if err = rows.Scan(&data.SenderId, &data.ReceiverId, &data.Status, &data.CreatedAt, &data.AmountCents, &data.Description); err != nil {
			return nil, 0, s.wrapScanError(err)
		}
		accountTransactionsData = append(accountTransactionsData, data)
	}

	const queryTotal = `SELECT COUNT("id") FROM transactions WHERE "senderId" = @accountId OR "receiverId" = @accountId`
	row := s.db.QueryRowContext(ctx, queryTotal, pgx.NamedArgs{
		"accountId": accountId,
		"limit":     limit,
		"offset":    offset,
	})
	if err = row.Err(); err != nil {
		return nil, 0, s.wrapQueryError(err)
	}
	var total int64
	err = row.Scan(&total)
	if err != nil {
		return nil, 0, s.wrapQueryError(err)
	}

	return accountTransactionsData, total, nil
}

func (s *Service) GetAccountDataById(ctx context.Context, senderId int64) (web.UserAccountData, error) {
	const accountQuery = `SELECT accounts."balanceCents", accounts."status", COALESCE("accountOwners"."userId", 0) FROM accounts 
    LEFT JOIN "accountOwners" ON accounts."ownerId" = "accountOwners".id WHERE accounts."id" = $1`
	row := s.db.QueryRowContext(ctx, accountQuery, senderId)
	if err := row.Err(); err != nil {
		return web.UserAccountData{}, s.wrapQueryError(err)
	}

	var userAccountData web.UserAccountData
	if err := row.Scan(&userAccountData.BalanceCents, &userAccountData.Status, &userAccountData.UserId); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return web.UserAccountData{}, cerrors.NewErrorWithUserMessage(ercodes.AccountDoesntExist, err,
				"Счёта, указанного в транзакции, не существует")
		}
		return web.UserAccountData{}, s.wrapScanError(err)
	}
	return userAccountData, nil
}

func (s *Service) CreateTransaction(ctx context.Context, senderId, receiverId, amountCents int64, description string) (int64, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, s.wrapQueryError(err)
	}

	defer func() {
		if tempErr := tx.Rollback(); tempErr != nil {
			err = s.wrapQueryError(tempErr)
		}
	}()

	errCh := make(chan error, 1)
	defer close(errCh)
	idCh := make(chan int64, 1)
	defer close(idCh)

	go func() {
		const queryTransaction = `INSERT INTO transactions ("senderId", "receiverId", "amountCents", description) VALUES (@senderId, @receiverId, @amountCents, @description) RETURNING id`
		var id int64
		err := tx.QueryRow(queryTransaction, pgx.NamedArgs{
			"senderId":    senderId,
			"receiverId":  receiverId,
			"amountCents": amountCents,
			"description": description,
		}).Scan(&id)
		idCh <- id
		errCh <- err
	}()

	go func() {
		const querySenderUpdate = `UPDATE accounts SET "balanceCents" = "balanceCents" - @amountCents WHERE id = @senderId`
		_, err := tx.ExecContext(ctx, querySenderUpdate, pgx.NamedArgs{
			"amountCents": amountCents,
			"senderId":    senderId,
		})
		errCh <- err
	}()

	id := <-idCh
	for i := 0; i < 2; i++ {
		if err = <-errCh; err != nil {
			return 0, s.wrapQueryError(err)
		}
	}

	if err = tx.Commit(); err != nil {
		return 0, s.wrapQueryError(err)
	}

	return id, err
}

func (s *Service) GetAtmDataByLogin(ctx context.Context, login string) (web.AtmData, error) {
	const query = `SELECT atms.id, atms.password, atms."cashCents", accounts.id as "hasPersonalData"
				   FROM atms
				   INNER JOIN "accountOwners" ON atms.id = "accountOwners"."atmId" 
					INNER JOIN "accounts" ON "accountOwners".id = "accounts"."ownerId"
				   WHERE atms.login = @login`

	row := s.db.QueryRowContext(ctx, query,
		pgx.NamedArgs{
			"login": login,
		},
	)
	if err := row.Err(); err != nil {
		return web.AtmData{}, s.wrapQueryError(err)
	}

	var atmData web.AtmData
	if err := row.Scan(&atmData.Id, &atmData.PasswordHash, &atmData.CashCents, &atmData.AccountId); err != nil {
		return web.AtmData{}, s.wrapScanError(err)
	}
	return atmData, nil
}

func (s *Service) UpdateAtmCash(ctx context.Context, amountCents, atmId int64) error {
	const query = `UPDATE atms SET "cashCents" = "cashCents" + @amountCents WHERE id = @atmId`

	_, err := s.db.ExecContext(ctx, query, pgx.NamedArgs{
		"amountCents": amountCents,
		"atmId":       atmId,
	})
	if err != nil {
		return s.wrapQueryError(err)
	}
	return nil
}

func (s *Service) UpdateAtmAccount(ctx context.Context, amountCents, accountId int64) error {
	const query = `UPDATE accounts SET "balanceCents" = "balanceCents" + @amountCents WHERE id = @accountId`

	_, err := s.db.ExecContext(ctx, query, pgx.NamedArgs{
		"amountCents": amountCents,
		"accountId":   accountId,
	})
	if err != nil {
		return s.wrapQueryError(err)
	}
	return nil
}

func (s *Service) LogCashOperation(ctx context.Context, atmId, amountCents, userAccountId int64) error {
	var query string
	if userAccountId == 0 {
		query = `INSERT INTO "cashOperations" ("atmAccountId", "amountCents") VALUES (@atmId, @amountCents)`
	} else {
		query = `INSERT INTO "cashOperations" ("atmAccountId", "userAccountId", "amountCents") VALUES (@atmId, @userAccountId, @amountCents)`
	}

	_, err := s.db.ExecContext(ctx, query, pgx.NamedArgs{
		"atmId":         atmId,
		"amountCents":   amountCents,
		"userAccountId": userAccountId,
	})
	if err != nil {
		return s.wrapQueryError(err)
	}
	return nil
}

func (s *Service) ConfirmTransaction(ctx context.Context, confirmationTime time.Duration) error {
	const queryTransactions = `SELECT "id", "senderId", "receiverId", "amountCents" FROM transactions WHERE current_timestamp - "createdAt" >= @confirmationTime AND status = 'BLOCKED'`
	rows, err := s.db.QueryContext(ctx, queryTransactions,
		pgx.NamedArgs{
			"confirmationTime": confirmationTime,
		},
	)
	if err != nil {
		return s.wrapQueryError(err)
	}

	var transactionsToApply []transaction_manager.TransactionToApply
	for rows.Next() {
		var data transaction_manager.TransactionToApply
		if err = rows.Scan(&data.Id, &data.SenderId, &data.ReceiverId, &data.AmountCents); err != nil {
			return s.wrapScanError(err)
		}
		transactionsToApply = append(transactionsToApply, data)
	}

	for _, transaction := range transactionsToApply {
		if err = s.applyTransaction(ctx, transaction); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) ConfirmTransactionById(ctx context.Context, transactionId int64) error {
	const queryTransactions = `
SELECT "id", "senderId", "receiverId", "amountCents" 
FROM transactions 
WHERE id = $1 AND status = 'BLOCKED'`

	var tr transaction_manager.TransactionToApply
	err := s.db.QueryRow(queryTransactions, transactionId).Scan(&tr.Id, &tr.SenderId, &tr.ReceiverId, &tr.AmountCents)
	if err != nil {
		return s.wrapQueryError(err)
	}
	return s.applyTransaction(ctx, tr)
}

func (s *Service) applyTransaction(ctx context.Context, transaction transaction_manager.TransactionToApply) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return s.wrapQueryError(err)
	}

	defer func() {
		if tempErr := tx.Rollback(); tempErr != nil {
			err = s.wrapQueryError(tempErr)
		}
	}()

	const queryTransaction = `UPDATE transactions SET status = 'CONFIRMED' WHERE id = $1`
	_, err = tx.ExecContext(ctx, queryTransaction, transaction.Id)
	if err != nil {
		return s.wrapQueryError(err)
	}

	const queryReceiverUpdate = `UPDATE accounts SET "balanceCents" = "balanceCents" + @amountCents WHERE id = @receiverId`

	_, err = tx.ExecContext(ctx, queryReceiverUpdate, pgx.NamedArgs{
		"amountCents": transaction.AmountCents,
		"receiverId":  transaction.ReceiverId,
	})
	if err != nil {
		return s.wrapQueryError(err)
	}
	if err = tx.Commit(); err != nil {
		return s.wrapQueryError(err)
	}
	return nil
}
