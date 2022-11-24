package rabbitmq

import (
	"log"
	"rmq_service/config"
	"rmq_service/pkg/logger"
	"rmq_service/pkg/rabbitmq"
	"time"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/streadway/amqp"
)

var (
	publishedMessages = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "email_published_rabbitmq_messages_total",
			Help: "The total number of published RabbitMQ messages",
		},
	)
)

// Emils rabbitmq publisher
type EmailsPublisher struct {
	amqpChan 	*amqp.Channel
	cfg 			*config.Config
	logger 		logger.Logger
}

// Emails rabbitmq publisher constructor
func NewEmailsPublisher(cfg *config.Config, logger logger.Logger) (*EmailsPublisher, error) {
	mqConn, err := rabbitmq.NewRabbitMQConn(cfg)
	if err != nil {
		return nil, err
	}

	amqpChan, err := mqConn.Channel()
	if err != nil {
		log.Fatalf("Channel(): %s", err)
		return nil, err
	}

	return &EmailsPublisher{amqpChan: amqpChan, cfg: cfg, logger: logger}, nil
}

// Create exchange and queue
func (p *EmailsPublisher) SetupExchangeAndQueue(exchange, queueName, bindingKey, consumerTag string) error {
	p.logger.Infof("Declaring exchange: %s", exchange)
	err := p.amqpChan.ExchangeDeclare(
		exchange,
		exchangeKind,
		exchangeDurable,
		exchangeAutoDelete,
		exchangeInternal,
		exchangeNoWait,
		nil,
	)

	if err != nil {
		log.Fatalf("ExchageDeclare(): %s", err)
		return err
	}

	queue, err := p.amqpChan.QueueDeclare(
		queueName,
		queueDurable,
		queueAutoDelete,
		queueExclusive,
		queueNoWait,
		nil,
	)

	if err != nil {
		log.Fatalf("QueueDeclare(): %s", err)
		return err
	}

	p.logger.Infof("Declared queue, binding it to exchange:\n Queue: %v\nMessage Count: %v\n" + 
			"Consumer Count: %v\nExchange: %v\nBinding Key: %v\n",
			queue.Name,
			queue.Messages,
			queue.Consumers,
			exchange,
			bindingKey,
		)

	err = p.amqpChan.QueueBind(
		queue.Name,
		bindingKey,
		exchange,
		queueNoWait,
		nil,
	)

	if err != nil {
		log.Fatalf("QueueBind(): %s", err)
		return err
	}

	p.logger.Infof("Queue bound to exchange, starting to consume from queue, consumerTag: %v", consumerTag)

	return nil
}

// Close messages chan
func (p *EmailsPublisher) CloseChan() {
	if err := p.amqpChan.Close(); err != nil {
		p.logger.Errorf("EmailsPublisher::CloseChan(): %v", err)
	}
}

// Publish message
func (p *EmailsPublisher) Publish(body []byte, contentType string) error {
	p.logger.Infof("Pulishing message Exchange: %s, RoutingKey: %s", p.cfg.RabbitMQ.Exchange, p.cfg.RabbitMQ.RoutingKey)
	
	if err := p.amqpChan.Publish(
		p.cfg.RabbitMQ.Exchange,
		p.cfg.RabbitMQ.RoutingKey,
		publishMandatory,
		publishImmediate,
		amqp.Publishing{
			ContentType: contentType,
			DeliveryMode: amqp.Persistent,
			MessageId: uuid.New().String(),
			Timestamp: time.Now(),
			Body: body,
		},
	); err != nil {
		log.Fatalf("Publish(): %s", err)
		return err
	}

	publishedMessages.Inc()
	return nil
}
