package random

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"x-bank-ms-bank/cerrors"
	"x-bank-ms-bank/ercodes"
)

type Service struct {
}

func NewService() Service {
	return Service{}
}

func (s *Service) GenerateString(_ context.Context, set string, size int) (string, error) {
	randomBytes := make([]byte, 2)
	res := make([]byte, size)
	for i := 0; i < size; i++ {
		randomNum, err := s.GenerateRandomNum(randomBytes)
		if err != nil {
			return "", err
		}

		res[i] = set[int(randomNum)%len(set)]
	}

	return string(res), nil
}

func (s *Service) GenerateRandomNum(buf []byte) (uint16, error) {
	_, err := rand.Read(buf)
	if err != nil {
		return 0, cerrors.NewErrorWithUserMessage(ercodes.RandomGeneration, err, "Ошибка генерации случайного числа")
	}

	num := binary.BigEndian.Uint16(buf)

	return num, nil
}
