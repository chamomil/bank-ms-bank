package web

import (
	"context"
	"x-bank-ms-bank/entity"
)

type (
	AccountStorage interface {
		GetUserAccounts(ctx context.Context, userId int64) ([]entity.UserAccountData, error)
		OpenUserAccount(ctx context.Context, userId int64) error
		BlockUserAccount(ctx context.Context, accountId int64) error
		GetAccountHistory(ctx context.Context, accountId, limit, offset int64) ([]entity.AccountTransactionsData, int64, error)
		UpdateAtmAccount(ctx context.Context, amountCents, accountId int64) error
		GetAccountDataById(ctx context.Context, senderId int64) (entity.UserAccountData, error)
	}

	TransactionStorage interface {
		CreateTransaction(ctx context.Context, senderId, receiverId, amountCents int64, description string) (int64, error)
		ConfirmTransactionById(ctx context.Context, transactionId int64) error
	}

	AtmStorage interface {
		GetAtmDataByLogin(ctx context.Context, login string) (entity.AtmData, error)
		UpdateAtmCash(ctx context.Context, amountCents, atmId int64) error
		LogCashOperation(ctx context.Context, atmId, amountCents, userId int64) error
	}

	PasswordHasher interface {
		CompareHashAndPassword(ctx context.Context, password string, hashedPassword []byte) error
		HashPassword(_ context.Context, password []byte, cost int) ([]byte, error)
	}
)
