package server

import (
	"rmq_service/config"
	"rmq_service/pkg/logger"

	"rmq_service/internal/email/delivery/rabbitmq"
	"rmq_service/internal/interceptors"
	"rmq_service/pkg/metrics"

	"github.com/jmoiron/sqlx"
	"github.com/streadway/amqp"
	"gopkg.in/gomail.v2"
)

// Server
type Server struct {
	db					*sqlx.DB
	mailDialer	*gomail.Dialer
	amqpConn		*amqp.Connection
	logger 			logger.Logger
	cfg 				*config.Config
}

// Server constructor
func NewEmailServer(
	amqpConn *amqp.Connection,
	logger logger.Logger,
	cfg *config.Config,
	mailDialer *gomail.Dialer,
	db *sqlx.DB) *Server {
		return &Server{
			db: db,
			amqpConn: amqpConn,
			logger: logger,
			mailDialer: mailDialer,
			cfg:	cfg,
		}
}

// Run server
func (s *Server) Run() error {
	metric, err := metrics.CreateMetrics(s.cfg.Metrics.URL, s.cfg.Metrics.ServiceName)
	if err != nil {
		s.logger.Errorf("CreateMetrics Error: %s", err)
	}

	s.logger.Info(
		"Metrics available URL: %s, ServiceName: %s",
		s.cfg.Metrics.URL,
		s.cfg.Metrics.ServiceName,
	)

	emailsPublisher, err := rabbitmq.NewEmailsPublisher(s.cfg, s.logger)
	if err != nil {
		return err
	}
	defer emailsPublisher.CloseChan()
	s.logger.Info("Emails Publisher initialized")

	im := interceptors.NewInterceptorManager(s.logger, s.cfg, metric)

	emailRepository := repository.NewEmailsRepository(s.db)
	
	
	return nil
}
