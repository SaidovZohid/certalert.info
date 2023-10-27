package main

import (
	"context"
	"fmt"
	"os"

	"github.com/SaidovZohid/certalert.info/api"
	"github.com/SaidovZohid/certalert.info/config"
	"github.com/SaidovZohid/certalert.info/pkg/logger"
	"github.com/SaidovZohid/certalert.info/storage"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/redis/go-redis/v9"
)

func main() {
	logger.Init()
	log := logger.GetLogger()
	log.Info("logger initialized")

	cfg := config.Load()
	log.Info("config initialized")
	databaseUrl := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.Postgres.Host,
		cfg.Postgres.Port,
		cfg.Postgres.User,
		cfg.Postgres.Password,
		cfg.Postgres.Database,
	)

	// psqlConn, err := sqlx.Connect("postgres", psqlUrl)
	// if err != nil {
	// 	log.Fatalf("failed to connect to database: %v", err)
	// }

	// this returns connection pool
	dbPool, err := pgxpool.Connect(context.Background(), databaseUrl)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer dbPool.Close()
	// defer func() {
	// 	if err := psqlConn.Close(); err != nil {
	// 		log.Fatalf("ERROR while closing connection: %v", err)
	// 	}
	// }()

	rdb := redis.NewClient(&redis.Options{
		Addr: cfg.Redis,
	})

	strg := storage.NewStoragePg(dbPool, log)
	inMemory := storage.NewInMemoryStorage(rdb)

	app := api.New(&api.RoutetOptions{
		Cfg:      &cfg,
		Log:      log,
		Strg:     strg,
		InMemory: inMemory,
	})

	log.Info("HTTP running in PORT -> ", cfg.HttpPort)
	log.Fatal("error while listening http port:", app.Listen(cfg.HttpPort))
}
