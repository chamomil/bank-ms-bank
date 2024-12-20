package jwt

import (
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"os"
	"strings"
	"time"
	"x-bank-ms-bank/auth"
	"x-bank-ms-bank/cerrors"
	"x-bank-ms-bank/ercodes"
)

type (
	RS256 struct {
		PrivateKey *rsa.PrivateKey
		PublicKey  *rsa.PublicKey
	}
)

func NewRS256(pathPrivateKey, pathPublicKey string) (RS256, error) {
	privateKey, err := func(path string) (*rsa.PrivateKey, error) {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}

		block, _ := pem.Decode(data)
		if block == nil || block.Type != "RSA PRIVATE KEY" {
			return nil, errors.New("Ошибка парсинга ключа")

		}

		key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}

		return key, nil
	}(pathPrivateKey)

	if err != nil {
		return RS256{}, err
	}

	publicKey, err := func(path string) (*rsa.PublicKey, error) {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}

		block, _ := pem.Decode(data)
		if block == nil || block.Type != "PUBLIC KEY" {
			return nil, errors.New("Ошибка парсинга ключа")
		}

		key, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			return nil, err
		}

		rsaPublicKey, ok := key.(*rsa.PublicKey)
		if !ok {
			return nil, errors.New("Ошибка преобразования в rsa.PublicKey")
		}

		return rsaPublicKey, nil
	}(pathPublicKey)

	if err != nil {
		return RS256{}, err
	}

	return RS256{
		PrivateKey: privateKey,
		PublicKey:  publicKey,
	}, nil
}

func (R *RS256) Authorize(ctx context.Context, claims auth.Claims) ([]byte, error) {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"RS256","typ":"JWT"}`))

	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return nil, cerrors.NewErrorWithUserMessage(ercodes.RS256Authorization, err, "Ошибка преобразования payload")
	}
	payload := base64.RawURLEncoding.EncodeToString(claimsJSON)

	signData := header + "." + payload
	hashed := sha256.Sum256([]byte(signData))

	signature, err := rsa.SignPKCS1v15(nil, R.PrivateKey, crypto.SHA256, hashed[:])
	if err != nil {
		return nil, cerrors.NewErrorWithUserMessage(ercodes.RS256Authorization, err, "Ошибка при подписывании токена")
	}

	token := header + "." + payload + "." + base64.RawURLEncoding.EncodeToString(signature)

	return []byte(token), nil
}

func (R *RS256) VerifyAuthorization(ctx context.Context, authorization []byte) (auth.Claims, error) {
	data := strings.Split(string(authorization), ".")
	if len(data) != 3 {
		return auth.Claims{}, errors.New("Токен не валиден")
	}

	signData := data[0] + "." + data[1]

	providedSignature, err := base64.RawURLEncoding.DecodeString(data[2])
	if err != nil {
		return auth.Claims{}, cerrors.NewErrorWithUserMessage(ercodes.RS256Authorization, err, "Ошибка преобразования подписи")
	}

	hashed := sha256.Sum256([]byte(signData))

	err = rsa.VerifyPKCS1v15(R.PublicKey, crypto.SHA256, hashed[:], providedSignature)
	if err != nil {
		return auth.Claims{}, cerrors.NewErrorWithUserMessage(ercodes.RS256Authorization, err, "Токен не валиден")
	}

	claimsJSON, err := base64.RawURLEncoding.DecodeString(data[1])
	if err != nil {
		return auth.Claims{}, cerrors.NewErrorWithUserMessage(ercodes.RS256Authorization, err, "Ошибка преобразования payload")
	}

	var claims auth.Claims
	err = json.Unmarshal(claimsJSON, &claims)
	if err != nil {
		return auth.Claims{}, cerrors.NewErrorWithUserMessage(ercodes.RS256Authorization, err, "Данные авторизации не соответствуют шаблону")
	}

	if claims.ExpiresAt < time.Now().Unix() {
		return auth.Claims{}, cerrors.NewErrorWithUserMessage(ercodes.RS256Authorization, errors.New("Время жизни токена истекло"), "Время жизни токена истекло")
	}

	return claims, nil
}
