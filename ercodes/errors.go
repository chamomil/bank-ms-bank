package ercodes

import "x-bank-ms-bank/cerrors"

const (
	_ cerrors.Code = -iota

	RandomGeneration
	BcryptHashing
	HS512Authorization
	RS256Authorization
	PostgresQuery
	PostgresScan
	BlockedAccount
	NotEnoughMoney
	WrongPassword
	AccessDenied
	AccountDoesntExist
	InvalidStatus
)
