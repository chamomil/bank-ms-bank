package hasher

import (
	"context"
	"golang.org/x/crypto/bcrypt"
	"x-bank-ms-bank/cerrors"
	"x-bank-ms-bank/ercodes"
)

type (
	Service struct {
	}
)

func NewService() Service {
	return Service{}
}

func (s *Service) HashPassword(_ context.Context, password []byte, cost int) ([]byte, error) {
	passwordHash, err := bcrypt.GenerateFromPassword(password, cost)
	if err != nil {
		return nil, cerrors.NewErrorWithUserMessage(ercodes.BcryptHashing, err, "Ошибка хэширования пароля")
	}

	return passwordHash, nil
}

func (s *Service) CompareHashAndPassword(_ context.Context, password string, hashedPassword []byte) error {
	err := bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
	if err != nil {
		return cerrors.NewErrorWithUserMessage(ercodes.WrongPassword, err, "Неверный логин или пароль")
	}
	return nil
}
