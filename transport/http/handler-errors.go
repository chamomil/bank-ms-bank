package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"x-bank-ms-bank/cerrors"
)

type (
	TransportError struct {
		InternalCode string `json:"internalCode"`
		DevMessage   string `json:"devMessage"`
		UserMessage  string `json:"userMessage"`
	}

	errorHandler struct {
		defaultStatusCode int
		statusCodes       map[cerrors.Code]int
	}
)

func errorMessage(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func (h *errorHandler) setTransportError(w http.ResponseWriter, transportError TransportError, statusCode int) {
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(&transportError)
}

func (h *errorHandler) setError(w http.ResponseWriter, err error) {
	var cErr *cerrors.Error
	if !errors.As(err, &cErr) {
		h.setTransportError(w, TransportError{
			DevMessage:  errorMessage(err),
			UserMessage: "Неизвестная ошибка",
		}, http.StatusBadRequest)
		return
	}

	statusCode, ok := h.statusCodes[cErr.Code]
	if !ok {
		statusCode = h.defaultStatusCode
	}

	h.setTransportError(w, TransportError{
		InternalCode: strconv.FormatInt(int64(cErr.Code), 10),
		DevMessage:   errorMessage(cErr.Origin),
		UserMessage:  cErr.UserMessage,
	}, statusCode)
}

func (h *errorHandler) setBadRequestError(w http.ResponseWriter, err error) {
	h.setTransportError(w, TransportError{
		DevMessage: errorMessage(err), UserMessage: "Ошибка запроса",
	}, http.StatusBadRequest)
}

func (h *errorHandler) setMethodNotAllowedError(w http.ResponseWriter) {
	h.setTransportError(w, TransportError{
		UserMessage: "Метод не поддерживается",
	}, http.StatusMethodNotAllowed)
}

func (h *errorHandler) setNotFoundError(w http.ResponseWriter) {
	h.setTransportError(w, TransportError{
		UserMessage: "Не найдено",
	}, http.StatusNotFound)
}

func (h *errorHandler) setUnauthorizedError(w http.ResponseWriter, err error) {
	h.setTransportError(w, TransportError{
		DevMessage: errorMessage(err), UserMessage: "Не авторизован",
	}, http.StatusUnauthorized)
}

func (h *errorHandler) setUnprocessableEntityError(w http.ResponseWriter, ve validationErrors) {
	w.WriteHeader(http.StatusUnprocessableEntity)
	_ = json.NewEncoder(w).Encode(&ve)
}

func (h *errorHandler) setFatalError(w http.ResponseWriter, v interface{}) {
	var devMessage string

	switch T := v.(type) {
	case error:
		devMessage = T.Error()
	case string:
		devMessage = T
	default:
		devMessage = "UNKNOWN FATAL ERROR"
	}

	h.setTransportError(w, TransportError{
		DevMessage:  devMessage,
		UserMessage: "Фатальная ошибка",
	}, http.StatusInternalServerError)
}
