package http

import (
	"net/http"
	"regexp"
)

var (
	isValidLogin = regexp.MustCompile("^[a-z0-9_-]{6,32}$").MatchString
)

func (t *Transport) validate(w http.ResponseWriter, v validatable) bool {
	ve := v.validate()
	if len(ve) > 0 {
		t.errorHandler.setUnprocessableEntityError(w, ve)
		return false
	}

	return true
}

func (u *ATMAuthData) validate() (ve validationErrors) {
	ve = make(validationErrors, 0, 2)

	if !isValidLogin(u.Login) {
		ve.Add("Неверный логин")
	}

	//if len(u.Password) < 6 || len(u.Password) > 16 {
	//	ve.Add("Неверный пароль")
	//}

	return
}

func (u *ATMOperationData) validate() (ve validationErrors) {
	ve = make(validationErrors, 0, 1)

	if u.AmountCents <= 0 {
		ve.Add("Неверная сумма для перевода")
	}

	return
}

func (u *ATMUserOperationData) validate() (ve validationErrors) {
	ve = make(validationErrors, 0, 2)

	if u.AmountCents <= 0 {
		ve.Add("Неверная сумма для перевода")
	}
	if u.AccountId < 0 {
		ve.Add("Неверный id для транзакции")
	}

	return
}

func (u *TransactionData) validate() (ve validationErrors) {
	ve = make(validationErrors, 0, 2)

	if u.AmountCents <= 0 {
		ve.Add("Неверная сумма для перевода")
	}

	if u.SenderId < 0 || u.ReceiverId < 0 || u.SenderId == u.ReceiverId {
		ve.Add("Неверный id для транзакции")
	}

	return
}
