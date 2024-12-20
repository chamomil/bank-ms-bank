package http

import "net/http"

func (t *Transport) routes() http.Handler {
	corsHandler := t.corsHandler("*", "*", "*", "")
	corsMiddleware := t.corsMiddleware(corsHandler)

	defaultMiddlewareGroup := middlewareGroup{
		t.panicMiddleware,
		corsMiddleware,
	}

	userMiddlewareGroup := middlewareGroup{
		t.panicMiddleware,
		corsMiddleware,
		t.authMiddleware(false),
	}

	ATMMiddlewareGroup := middlewareGroup{
		t.panicMiddleware,
		corsMiddleware,
		t.basicAuthMiddleware(),
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/", defaultMiddlewareGroup.Apply(t.handlerNotFound))
	mux.HandleFunc("GET /v1/me/accounts", userMiddlewareGroup.Apply(t.handlerUserAccounts))

	mux.HandleFunc("POST /v1/accounts", userMiddlewareGroup.Apply(t.handlerOpenAccount))
	mux.HandleFunc("PATCH /v1/accounts/{accountId}", userMiddlewareGroup.Apply(t.handlerBlockAccount))
	mux.HandleFunc("GET /v1/accounts/{accountId}/history", userMiddlewareGroup.Apply(t.handlerAccountHistory))

	mux.HandleFunc("POST /v1/transactions", userMiddlewareGroup.Apply(t.handlerAccountTransaction))
	mux.HandleFunc("PATCH /v1/transactions/{id}", userMiddlewareGroup.Apply(t.handlerChangeTransactionStatus))
	mux.HandleFunc("POST /v1/atm/supplement", ATMMiddlewareGroup.Apply(t.handlerATMSupplement))
	mux.HandleFunc("POST /v1/atm/withdrawal", ATMMiddlewareGroup.Apply(t.handlerATMWithdrawal))
	mux.HandleFunc("POST /v1/atm/user/supplement", ATMMiddlewareGroup.Apply(t.handlerATMUserSupplement))
	mux.HandleFunc("POST /v1/atm/user/withdrawal", ATMMiddlewareGroup.Apply(t.handlerATMUserWithdrawal))

	return mux
}
