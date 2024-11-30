package transaction_manager

import (
	"context"
	"time"
)

type (
	Service struct {
		transactionStorage TransactionStorage
	}
)

func NewService(transactionStorage TransactionStorage) Service {
	return Service{
		transactionStorage: transactionStorage,
	}
}

const (
	confirmationTime = 5 * time.Minute
)

func (s *Service) ApplyTransactions(ctx context.Context) error {
	return s.transactionStorage.ConfirmTransaction(ctx, confirmationTime)
}
