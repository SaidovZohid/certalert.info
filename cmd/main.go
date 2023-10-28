package main

import (
	"context"
	"fmt"
	"os"

	"github.com/SaidovZohid/certalert.info/api"
	"github.com/SaidovZohid/certalert.info/config"
	"github.com/SaidovZohid/certalert.info/pkg/logger"
	"github.com/SaidovZohid/certalert.info/pkg/utils"
	"github.com/SaidovZohid/certalert.info/storage"
	"github.com/SaidovZohid/certalert.info/telegram"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
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

	// this returns connection pool
	dbPool, err := pgxpool.Connect(context.Background(), databaseUrl)
	if err != nil {
		log.Error("Unable to connect to database: " + err.Error())
		os.Exit(1)
	}
	defer dbPool.Close()

	bot, err := tgbotapi.NewBotAPI(cfg.TelegramApiToken)
	if err != nil {
		log.Fatalf("Failed to make new bot api: %v", err)
	}

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

	go func() {
		log.Info("Initializing regular domain information update...")

		// Initiate the function to update domain information regularly
		utils.UpdateDomainInformationRegularly(context.Background(), strg, log, &cfg)
	}()
	go func() {
		log.Info("Initializing and starting the Telegram bot...")

		telegramBot := telegram.NewBot(bot, &log, &cfg)

		if err := telegramBot.Start(); err != nil {
			log.Fatal(fmt.Sprintf("Error while starting bot: %v", err))
		}
	}()

	log.Info("HTTP running in PORT -> ", cfg.HttpPort)
	log.Fatal("Error while listening http port:", app.Listen(cfg.HttpPort))
}
