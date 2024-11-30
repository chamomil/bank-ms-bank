package web

import "time"

type (
	UserAccountData struct {
		Id           int64
		BalanceCents int64
		Status       string
		UserId       int64
	}

	AccountTransactionsData struct {
		SenderId    int64
		ReceiverId  int64
		Status      string
		CreatedAt   time.Time
		AmountCents int64
		Description string
	}

	AtmData struct {
		Id           int64
		AccountId    int64
		PasswordHash []byte
		CashCents    int64
	}
)
