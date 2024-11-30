package transaction_manager

import (
	"context"
	"time"
)

type (
	TransactionStorage interface {
		ConfirmTransaction(ctx context.Context, confirmationTime time.Duration) error
	}
)
