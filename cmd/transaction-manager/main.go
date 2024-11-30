package main

import (
	"context"
	"flag"
	"log"
	"x-bank-ms-bank/config"
	transaction_manager "x-bank-ms-bank/core/transaction-manager"
	"x-bank-ms-bank/infra/postgres"
)

var (
	configFile = flag.String("config", "config.json", "")
)

func main() {
	conf, err := config.Read(*configFile)
	if err != nil {
		log.Fatal(err)
	}
	postgresService, err := postgres.NewService(conf.Postgres.Login, conf.Postgres.Password, conf.Postgres.Host, conf.Postgres.Port, conf.Postgres.DataBase, conf.Postgres.MaxCons)
	if err != nil {
		log.Fatal(err)
	}

	service := transaction_manager.NewService(&postgresService)
	if err = service.ApplyTransactions(context.Background()); err != nil {
		log.Fatal(err)
	}
}
