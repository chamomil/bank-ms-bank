package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"x-bank-ms-bank/auth"
)

const (
	maxLimit      = 100
	minLimit      = 0
	defaultLimit  = 20
	minOffset     = 0
	defaultOffset = 0
)

func (t *Transport) handlerNotFound(w http.ResponseWriter, _ *http.Request) {
	t.errorHandler.setNotFoundError(w)
}

func (t *Transport) handlerUserAccounts(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(t.claimsCtxKey).(*auth.Claims)
	if !ok {
		t.errorHandler.setError(w, errors.New("отсутствуют claims в контексте"))
		return
	}
	userId := claims.Sub
	data, err := t.service.GetAccounts(r.Context(), userId)
	if err != nil {
		t.errorHandler.setError(w, err)
		return
	}

	var response UserAccountsResponse
	if data != nil {
		for _, entry := range data {
			userAccountsItem := UserAccountsResponseItem{
				Id:           entry.Id,
				BalanceCents: entry.BalanceCents,
				Status:       entry.Status,
			}
			response.Items = append(response.Items, userAccountsItem)
		}
	} else {
		response.Items = nil
	}

	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(data)
	if err != nil {
		t.errorHandler.setError(w, err)
		return
	}
}

func (t *Transport) handlerOpenAccount(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(t.claimsCtxKey).(*auth.Claims)
	if !ok {
		t.errorHandler.setError(w, errors.New("отсутствуют claims в контексте"))
		return
	}
	userId := claims.Sub
	if err := t.service.OpenAccount(r.Context(), userId); err != nil {
		t.errorHandler.setError(w, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (t *Transport) handlerBlockAccount(w http.ResponseWriter, r *http.Request) {
	accountId, err := strconv.ParseInt(r.PathValue("accountId"), 10, 64)
	if err != nil {
		t.errorHandler.setBadRequestError(w, err)
	}
	claims, ok := r.Context().Value(t.claimsCtxKey).(*auth.Claims)
	if !ok {
		t.errorHandler.setError(w, errors.New("отсутствуют claims в контексте"))
		return
	}
	userId := claims.Sub

	if err := t.service.BlockAccount(r.Context(), accountId, userId); err != nil {
		t.errorHandler.setError(w, err)
	}
	w.WriteHeader(http.StatusOK)
}

func (t *Transport) handlerAccountHistory(w http.ResponseWriter, r *http.Request) {
	accountId, err := strconv.ParseInt(r.PathValue("accountId"), 10, 64)
	if err != nil {
		t.errorHandler.setBadRequestError(w, err)
	}
	limit, err := strconv.ParseInt(r.URL.Query().Get("limit"), 10, 64)
	if err != nil || limit < minLimit || limit > maxLimit {
		limit = defaultLimit
	}
	offset, err := strconv.ParseInt(r.URL.Query().Get("offset"), 10, 64)
	if err != nil || offset < 0 {
		offset = defaultOffset
	}

	claims, ok := r.Context().Value(t.claimsCtxKey).(*auth.Claims)
	if !ok {
		t.errorHandler.setError(w, errors.New("отсутствуют claims в контексте"))
		return
	}
	userId := claims.Sub

	data, total, err := t.service.GetAccountHistory(r.Context(), accountId, userId, limit, offset)
	if err != nil {
		t.errorHandler.setError(w, err)
		return
	}

	var response AccountsHistoryResponse
	if data != nil {
		for _, entry := range data {
			userAccountsItem := AccountsHistoryResponseItem{
				SenderId:    entry.SenderId,
				ReceiverId:  entry.ReceiverId,
				Status:      entry.Status,
				CreatedAt:   entry.CreatedAt.Format("2006.01.02 15:04:05"),
				AmountCents: entry.AmountCents,
				Description: entry.Description,
			}
			response.Items = append(response.Items, userAccountsItem)
		}
	} else {
		response.Items = nil
	}
	response.Total = total

	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		t.errorHandler.setError(w, err)
		return
	}
}

func (t *Transport) handlerAccountTransaction(w http.ResponseWriter, r *http.Request) {
	var transactionData TransactionData
	if err := json.NewDecoder(r.Body).Decode(&transactionData); err != nil {
		t.errorHandler.setBadRequestError(w, err)
		return
	}
	if !t.validate(w, &transactionData) {
		return
	}
	claims, ok := r.Context().Value(t.claimsCtxKey).(*auth.Claims)
	if !ok {
		t.errorHandler.setError(w, errors.New("отсутствуют claims в контексте"))
		return
	}
	userId := claims.Sub

	if err := t.service.MakeTransaction(r.Context(), transactionData.SenderId, transactionData.ReceiverId, transactionData.AmountCents, userId, transactionData.Description); err != nil {
		t.errorHandler.setError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (t *Transport) handlerATMSupplement(w http.ResponseWriter, r *http.Request) {
	var atmSupplementData ATMOperationData
	if err := json.NewDecoder(r.Body).Decode(&atmSupplementData); err != nil {
		t.errorHandler.setBadRequestError(w, err)
		return
	}
	if !t.validate(w, &atmSupplementData) {
		return
	}

	basic, ok := r.Context().Value(t.basicCtxKey).(ATMAuthData)
	if !ok {
		t.errorHandler.setError(w, errors.New("отсутствуют basic auth в контексте"))
		return
	}
	if !t.validate(w, &basic) {
		return
	}

	if err := t.service.ATMSupplement(r.Context(), basic.Login, basic.Password, atmSupplementData.AmountCents); err != nil {
		t.errorHandler.setError(w, err)
	}

	w.WriteHeader(http.StatusOK)
}

func (t *Transport) handlerATMWithdrawal(w http.ResponseWriter, r *http.Request) {
	var atmWithdrawalData ATMOperationData
	if err := json.NewDecoder(r.Body).Decode(&atmWithdrawalData); err != nil {
		t.errorHandler.setBadRequestError(w, err)
		return
	}
	if !t.validate(w, &atmWithdrawalData) {
		return
	}

	basic, ok := r.Context().Value(t.basicCtxKey).(ATMAuthData)
	if !ok {
		t.errorHandler.setError(w, errors.New("отсутствуют basic auth в контексте"))
		return
	}
	if !t.validate(w, &basic) {
		return
	}

	if err := t.service.ATMWithdrawal(r.Context(), basic.Login, basic.Password, atmWithdrawalData.AmountCents); err != nil {
		t.errorHandler.setError(w, err)
	}

	w.WriteHeader(http.StatusOK)
}

func (t *Transport) handlerATMUserSupplement(w http.ResponseWriter, r *http.Request) {
	var atmUserSupplementData ATMUserOperationData
	if err := json.NewDecoder(r.Body).Decode(&atmUserSupplementData); err != nil {
		t.errorHandler.setBadRequestError(w, err)
		return
	}
	if !t.validate(w, &atmUserSupplementData) {
		return
	}

	basic, ok := r.Context().Value(t.basicCtxKey).(ATMAuthData)
	if !ok {
		t.errorHandler.setError(w, errors.New("отсутствуют basic auth в контексте"))
		return
	}
	if !t.validate(w, &basic) {
		return
	}

	if err := t.service.ATMUserSupplement(r.Context(), basic.Login, basic.Password, atmUserSupplementData.AmountCents, atmUserSupplementData.AccountId, 0); err != nil {
		t.errorHandler.setError(w, err)
	}

	w.WriteHeader(http.StatusOK)
}

func (t *Transport) handlerATMUserWithdrawal(w http.ResponseWriter, r *http.Request) {
	var atmUserWithdrawalData ATMUserOperationData
	if err := json.NewDecoder(r.Body).Decode(&atmUserWithdrawalData); err != nil {
		t.errorHandler.setBadRequestError(w, err)
		return
	}

	if !t.validate(w, &atmUserWithdrawalData) {
		return
	}
	basic, ok := r.Context().Value(t.basicCtxKey).(ATMAuthData)
	if !ok {
		t.errorHandler.setError(w, errors.New("отсутствуют basic auth в контексте"))
		return
	}
	if !t.validate(w, &basic) {
		return
	}

	if err := t.service.ATMUserWithdrawal(r.Context(), basic.Login, basic.Password, atmUserWithdrawalData.AmountCents, atmUserWithdrawalData.AccountId, 0); err != nil {
		t.errorHandler.setError(w, err)
	}

	w.WriteHeader(http.StatusOK)
}
