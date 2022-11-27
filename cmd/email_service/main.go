package main

import (
	"log"
	"os"

	config "rmq_service/config"
	"rmq_service/internal/server"
	"rmq_service/pkg/jaeger"
	"rmq_service/pkg/logger"
	"rmq_service/pkg/mailer"
	"rmq_service/pkg/postgres"
	"rmq_service/pkg/rabbitmq"

	"github.com/opentracing/opentracing-go"
)

func main() {
	log.Println("Starting server")
	
	configPath := config.GetConfigPath(os.Getenv("config")) // export config='local'
	cfg, err   := config.GetConfig(configPath)
	if err != nil {
		log.Fatalf("Loading config: %v", err)
	}

	//log.Fatalf("CUPA APP_LOGGER: %#v", log)
	log.Printf(
		"AppVersion: %s, LogLevel: %s, Mode: %s, SSL: %v",
		cfg.Server.AppVersion,
		cfg.Logger.Level,
		cfg.Server.Mode,
		cfg.Server.SSL,
	)
	log.Printf("Success parsed config: %#v", cfg.Server.AppVersion)

	amqpConn, err := rabbitmq.NewRabbitMQConn(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer amqpConn.Close()

	psqlDB, err := postgres.NewPsqlDB(cfg)
	if err != nil {
		log.Fatalf("Postgresql init: %s", err)
	}
	defer psqlDB.Close()

	log.Printf("PostgreSQL connected: %#v", psqlDB.Stats())

	tracer, closer, err := jaeger.InitJaeger(cfg)
	if err != nil {
		log.Fatal("cannot crate tracer", err)
	}
	log.Println("Jaeger connected")

	opentracing.SetGlobalTracer(tracer)
	defer closer.Close()
	log.Println("Opentracing connected")

	mailDialer := mailer.NewMailDialer(cfg)
	log.Println("Mail dialer connected")

	s := server.NewEmailServer(amqpConn, logger.NewApiLogger(cfg), cfg, mailDialer, psqlDB)

	log.Fatal(s.Run())
}
