package jwt

import (
	"context"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"strings"
	"time"
	"x-bank-ms-bank/auth"
	"x-bank-ms-bank/cerrors"
	"x-bank-ms-bank/ercodes"
)

type (
	HS512 struct {
		secret []byte
	}
)

func NewHS512(secret string) (HS512, error) {
	hs512SecretKey, err := hex.DecodeString(secret)
	if err != nil {
		return HS512{}, err
	}
	return HS512{
		secret: hs512SecretKey,
	}, nil
}

func (R *HS512) Authorize(_ context.Context, claims auth.Claims) ([]byte, error) {
	mac := hmac.New(sha512.New, R.secret)
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS512","typ":"JWT"}`))

	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return nil, cerrors.NewErrorWithUserMessage(ercodes.HS512Authorization, err, "Ошибка преобразования payload")
	}
	payload := base64.RawURLEncoding.EncodeToString(claimsJSON)

	signData := header + "." + payload

	_, err = mac.Write([]byte(signData))
	if err != nil {
		return nil, cerrors.NewErrorWithUserMessage(ercodes.HS512Authorization, err, "Ошибка при подписывании токена")
	}
	token := signData + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return []byte(token), nil
}

func (R *HS512) VerifyAuthorization(_ context.Context, authorization []byte) (auth.Claims, error) {
	data := strings.Split(string(authorization), ".")

	if len(data) != 3 {
		return auth.Claims{}, cerrors.NewErrorWithUserMessage(ercodes.HS512Authorization, nil, "Токен не валиден")
	}

	mac := hmac.New(sha512.New, R.secret)
	signData := data[0] + "." + data[1]

	_, err := mac.Write([]byte(signData))
	if err != nil {
		return auth.Claims{}, cerrors.NewErrorWithUserMessage(ercodes.HS512Authorization, err, "Ошибка при подписывании токена")
	}
	signature := mac.Sum(nil)

	providedSignature, err := base64.RawURLEncoding.DecodeString(data[2])
	if err != nil {
		return auth.Claims{}, cerrors.NewErrorWithUserMessage(ercodes.HS512Authorization, err, "Ошибка преобразования подписи")
	}
	if !hmac.Equal(signature, providedSignature) {
		return auth.Claims{}, cerrors.NewErrorWithUserMessage(ercodes.HS512Authorization, err, "Токен не валиден")
	}
	claimsJSON, err := base64.RawURLEncoding.DecodeString(data[1])
	if err != nil {
		return auth.Claims{}, cerrors.NewErrorWithUserMessage(ercodes.HS512Authorization, err, "Ошибка преобразования payload")
	}

	var userClaims auth.Claims
	err = json.Unmarshal(claimsJSON, &userClaims)
	if err != nil {
		return auth.Claims{}, cerrors.NewErrorWithUserMessage(ercodes.HS512Authorization, err, "Данные авторизации не соответствуют шаблону")
	}

	if userClaims.ExpiresAt < time.Now().Unix() {
		return auth.Claims{}, cerrors.NewErrorWithUserMessage(ercodes.HS512Authorization, err, "Время жизни токена истекло")
	}

	return userClaims, nil
}
