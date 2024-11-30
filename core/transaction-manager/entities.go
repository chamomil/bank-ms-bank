package transaction_manager

type (
	TransactionToApply struct {
		Id          int64
		SenderId    int64
		ReceiverId  int64
		AmountCents int64
	}
)
