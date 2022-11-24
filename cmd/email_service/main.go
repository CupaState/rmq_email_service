package main

import (
	"log"
	"os"

	config "rmq_service/config"
	"rmq_service/internal/server"
	"rmq_service/pkg/jaeger"
	logger "rmq_service/pkg/logger"
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

	appLogger := logger.NewApiLogger(cfg)
	appLogger.InitLogger()
	appLogger.Infof(
		"AppVersion: %s, LogLevel: %s, Mode: %s, SSL: %v",
		cfg.Server.AppVersion,
		cfg.Logger.Level,
		cfg.Server.Mode,
		cfg.Server.SSL,
	)
	appLogger.Infof("Success parsed config: %#v", cfg.Server.AppVersion)

	amqpConn, err := rabbitmq.NewRabbitMQConn(cfg)
	if err != nil {
		appLogger.Fatal(err)
	}
	defer amqpConn.Close()

	psqlDB, err := postgres.NewPsqlDB(cfg)
	if err != nil {
		appLogger.Fatalf("Postgresql init: %s", err)
	}
	defer psqlDB.Close()

	appLogger.Infof("PostgreSQL connected: %#v", psqlDB.Stats())

	tracer, closer, err := jaeger.InitJaeger(cfg)
	if err != nil {
		appLogger.Fatal("cannot crate tracer", err)
	}
	appLogger.Info("Jaeger connected")

	opentracing.SetGlobalTracer(tracer)
	defer closer.Close()
	appLogger.Info("Opentracing connected")

	mailDialer := mailer.NewMailDialer(cfg)
	appLogger.Info("Mail dialer connected")

	s := server.NewEmailServer(amqpConn, appLogger, cfg, mailDialer, psqlDB)

	log.Fatal(s.Run())
}
