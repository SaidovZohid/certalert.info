package main

import (
	"fmt"

	"github.com/SaidovZohid/certalert.info/api"
	"github.com/SaidovZohid/certalert.info/config"
	"github.com/SaidovZohid/certalert.info/pkg/logger"
	"github.com/SaidovZohid/certalert.info/storage"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
)

func main() {
	logger.Init()
	log := logger.GetLogger()
	log.Info("logger initialized")

	cfg := config.Load()
	log.Info("config initialized")
	psqlUrl := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.Postgres.Host,
		cfg.Postgres.Port,
		cfg.Postgres.User,
		cfg.Postgres.Password,
		cfg.Postgres.Database,
	)

	psqlConn, err := sqlx.Connect("postgres", psqlUrl)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: cfg.Redis,
	})

	strg := storage.NewStoragePg(psqlConn)
	inMemory := storage.NewInMemoryStorage(rdb)

	app := api.New(&api.RoutetOptions{
		Cfg:  &cfg,
		Log:  log,
		Strg: strg,
		InMemory: inMemory,
	})

	log.Info("HTTP running in PORT -> ", cfg.HttpPort)
	log.Fatal("error while listening http port:", app.Listen(cfg.HttpPort))
}

// // Create a TLS configuration
// tlsConfig := &tls.Config{}

// // Establish a connection to the server
// conn, err := tls.Dial("tcp", "uzmovi.com:443", tlsConfig)
// if err != nil {
// 	fmt.Println("Failed to connect:", err)
// 	return
// }
// defer conn.Close()

// // Retrieve the peer certificate
// cert := conn.ConnectionState().PeerCertificates[0]

// // Extract and print the certificate details
// fmt.Println("Certificate Subject:", cert.Subject)
// fmt.Println("Certificate Issuer:", cert.Issuer)
// fmt.Println("Certificate Expiration Date:", cert.NotAfter)
// fmt.Println("Certificate DNS:", cert.DNSNames)
// fmt.Println("Certificate Signature:", cert.Issuer.Organization)
