package config

import (
	"encoding/json"
	"os"
)

type (
	Config struct {
		Hs512SecretKey  string   `json:"hs512SecretKey"`
		Rs256PrivateKey string   `json:"rs256PrivateKey"`
		Rs256PublicKey  string   `json:"rs256PublicKey"`
		Postgres        Postgres `json:"postgres"`
	}

	Postgres struct {
		Login    string `json:"login"`
		Password string `json:"password"`
		Host     string `json:"host"`
		Port     int    `json:"port"`
		DataBase string `json:"dataBase"`
		MaxCons  int    `json:"maxCons"`
	}
)

func Read(filename string) (Config, error) {
	f, err := os.Open(filename)
	if err != nil {
		return Config{}, err
	}
	defer func() { _ = f.Close() }()

	var config Config
	err = json.NewDecoder(f).Decode(&config)
	if err != nil {
		return Config{}, err
	}

	return config, nil
}
