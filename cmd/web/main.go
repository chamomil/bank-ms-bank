package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
	"x-bank-ms-bank/config"
	"x-bank-ms-bank/core/web"
	"x-bank-ms-bank/infra/hasher"
	"x-bank-ms-bank/infra/postgres"
	"x-bank-ms-bank/transport/http"
	"x-bank-ms-bank/transport/http/jwt"
)

var (
	addr       = flag.String("addr", ":8081", "")
	configFile = flag.String("config", "config.json", "")
)

func main() {
	flag.Parse()
	conf, err := config.Read(*configFile)
	if err != nil {
		log.Fatal(err)
	}

	jwtHs512, err := jwt.NewHS512(conf.Hs512SecretKey)
	if err != nil {
		log.Fatal(err)
	}
	postgresService, err := postgres.NewService(conf.Postgres.Login, conf.Postgres.Password, conf.Postgres.Host, conf.Postgres.Port, conf.Postgres.DataBase, conf.Postgres.MaxCons)
	if err != nil {
		log.Fatal(err)
	}
	passwordHasher := hasher.NewService()

	service := web.NewService(&postgresService, &passwordHasher, &postgresService, &postgresService)
	transport := http.NewTransport(service, &jwtHs512)

	errCh := transport.Start(*addr)
	interruptsCh := make(chan os.Signal, 1)
	signal.Notify(interruptsCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-errCh:
		log.Fatal(err)
	case <-interruptsCh:
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()
		err = transport.Stop(shutdownCtx)
		if err != nil {
			log.Fatal(err)
		}
	}
}
