package auth

import "context"

type (
	Claims struct {
		Id string `json:"jti"`

		IssuedAt  int64 `json:"iat"`
		ExpiresAt int64 `json:"exp"`

		Sub int64 `json:"sub"`

		Is2FAToken      bool `json:"2fa"`
		HasPersonalData bool `json:"idf"`
	}

	Authorizer interface {
		Authorize(ctx context.Context, claims Claims) ([]byte, error)
		VerifyAuthorization(ctx context.Context, authorization []byte) (Claims, error)
	}
)
